// interior_stream_concurrent_write_test.go — PINS THE MECHANISM behind the live-editor "bad
// frame length" desync that used to be reported on the interior stream (handleInteriorFd in
// runCommand.ts). See docs/interior-stream-framing.md for the investigation this test was
// originally the deliverable of, and its "Fix" section for what changed.
//
// The mechanism (now FIXED in production, not reproduced from it — see below): a
// gatecommon.DriveHeld goroutine used to call Out.PlaceDrivenAt -> flushSendEvent ->
// WriteEvents on the node's shared *interiorStream — the SAME *interiorStream instance the
// node's own Update-loop goroutine writes to via EmitHeldBead/EmitNodeBeads. Both went
// through writeInteriorStreamFrame, which used to issue TWO separate io.Writer.Write calls
// per frame ([len:u32] header, then payload) with nothing spanning them — two goroutines
// racing on one pipe fd could interleave those writes (A's header, then B's header+payload,
// then A's payload) and desync the reader's [len:u32] framing exactly as observed.
//
// PRODUCTION no longer constructs this sharing at all: every gatecommon.DriveHeld goroutine
// now gets its OWN dedicated fd (Buffer.StreamKindDrive, nodes/Wiring/build_args.go's
// BuildArgs.DriveOut) — see docs/interior-stream-framing.md's "Fix" section for why a
// per-goroutine fd was chosen over a lock (this codebase has none: check-no-network-
// locks.sh's allowlist is empty) or channel-routing. writeInteriorStreamFrame ALSO now
// folds the header and payload into ONE io.Writer.Write call (this file's other change),
// which closes the specific two-separate-Write-calls race even for a genuinely single
// writer (a short write, a signal).
//
// Because BOTH of production's own fixes (per-goroutine fd, single-Write framing) are now
// true, this test cannot reach the bug through any real call path — s.write/s.WriteEvents
// no longer issue two Writes, so calling them concurrently from two goroutines against one
// shared io.Writer no longer reproduces the desync (confirmed: it does not fail). So this
// test reconstructs the ORIGINAL two-Write-calls framing BY HAND (writeTwoCallFrame below,
// a frozen copy of writeInteriorStreamFrame's pre-fix shape) to pin what made the original
// bug possible: TWO GOROUTINES, ONE io.Writer, TWO SEPARATE Write() CALLS PER FRAME is
// unsafe, full stop — independent of which production code path used to produce that
// shape. If a future change reintroduces a two-call frame write (on a shared writer or
// not) or reintroduces sharing a stream across goroutines, this documents exactly why
// either one, on its own, is enough to break framing.
//
// This is a violation of the model's single-writer-per-fd invariant
// (memory/feedback_no_single_writer_bridge.md), not a framing-format bug — BuildInterior-
// StreamFrame/BuildEventsSection were independently checked and always produce a byte count
// that matches their own declared length (see docs/interior-stream-framing.md's point 3).
package Wiring

import (
	"encoding/binary"
	"os"
	"sync"
	"testing"
	"time"

	B "github.com/dtauraso/wirefold/Buffer"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// maxFrameBytesForTest mirrors runCommand.ts's MAX_FRAME_BYTES (1<<20) — the same bound
// the ext host's splitFrames uses to detect a desynced reader and stop the stream (the
// exact failure mode this test reproduces).
const maxFrameBytesForTest = 1 << 20

// toTestStreamEvents is a local copy of main.go's toStreamEvents — kept local rather than
// exported from main.go (a `package main` file) so this test can build a real
// Buffer.BuildInteriorStreamFrame-shaped frame without introducing a Wiring->main import
// cycle. Same field-for-field mapping.
func toTestStreamEvents(events []wire.RowEvent) []B.StreamEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]B.StreamEvent, len(events))
	for i, e := range events {
		out[i] = B.StreamEvent{
			Kind: B.KindID(e.Kind), NodeRow: e.NodeRow, PortRow: e.PortRow,
			TargetRow: e.TargetRow, TargetPortRow: e.TargetPortRow, EdgeRow: e.EdgeRow,
			Slot: e.Slot, Value: e.Value, Bead: uint32(e.Bead),
			BeadSteps: float32(e.BeadSteps), SimLatencyMs: float32(e.SimLatencyMs),
			X: float32(e.X), Y: float32(e.Y), Z: float32(e.Z), F: float32(e.F),
			Label: e.Label, Debug: e.Debug, Text: e.Text,
		}
	}
	return out
}

// writeTwoCallFrame is a FROZEN COPY of writeInteriorStreamFrame's PRE-FIX shape (two
// separate io.Writer.Write calls per frame: header, then payload) — this file's own
// deliberately-preserved reproduction of the exact byte-level hazard the production fix
// removed. It must NOT be "fixed" to match the current single-Write writeInteriorStream-
// Frame — doing so would defeat the point of this test (see this file's header comment).
func writeTwoCallFrame(out *os.File, frame []byte) {
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	_, _ = out.Write(hdr[:])
	_, _ = out.Write(frame)
}

// TestInteriorStreamTwoCallSharedWriterMechanismStillDesyncs pins the ORIGINAL
// violation's mechanism by hand (see this file's header comment): two goroutines sharing
// ONE pipe fd, each framing its own frame as TWO separate Write() calls (writeTwoCallFrame)
// — exactly what writeInteriorStreamFrame used to do, called from two goroutines that used
// to share one *interiorStream before gatecommon.DriveHeld got its own dedicated fd
// (Buffer.StreamKindDrive) and writeInteriorStreamFrame got folded into one Write() call.
// Neither fix is exercised here on purpose: this test proves the OLD shape was unsafe on
// its own terms, not that the new shape remains safe (the other headless/wiring tests in
// this package cover that side).
//
// PASSES when the desync it exists to pin actually occurs; FAILS when it does not — see
// the inverted assertion at the end of this function for why (a permanent member of
// `go test ./...` cannot be "expected to fail" the way this file's original standalone
// reproduction was).
func TestInteriorStreamTwoCallSharedWriterMechanismStillDesyncs(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	buildFrame := func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte {
		return B.BuildInteriorStreamFrame(tick, present, value, ox, oy, oz, toTestStreamEvents(events))
	}

	const iterations = 20000
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine A: the node's own Update loop calling EmitHeldBead on every new input
	// value (Pulse/node.go's consume()) — reframed here through writeTwoCallFrame
	// directly on the shared pipe (bypassing s.write, which no longer has this shape).
	go func() {
		defer wg.Done()
		present := []uint8{1, 0, 0, 0}
		value := []int32{7, 0, 0, 0}
		ox := []float32{0, 0, 0, 0}
		oy := []float32{0, 0, 0, 0}
		oz := []float32{0, 0, 0, 0}
		for i := 0; i < iterations; i++ {
			writeTwoCallFrame(w, buildFrame(uint32(i), present, value, ox, oy, oz, nil))
		}
	}()

	// Goroutine B: a gatecommon.DriveHeld drive goroutine calling flushSendEvent on
	// every bead placement — the SAME shared pipe fd, a DIFFERENT OS goroutine, also
	// framing via writeTwoCallFrame (the pre-fix shape).
	go func() {
		defer wg.Done()
		present := []uint8{1, 0, 0, 0}
		value := []int32{7, 0, 0, 0}
		ox := []float32{0, 0, 0, 0}
		oy := []float32{0, 0, 0, 0}
		oz := []float32{0, 0, 0, 0}
		for i := 0; i < iterations; i++ {
			events := []wire.RowEvent{{
				Kind: "send", NodeRow: 0, PortRow: -1,
				TargetRow: 1, TargetPortRow: -1, EdgeRow: -1, Value: int32(i),
			}}
			writeTwoCallFrame(w, buildFrame(uint32(i), present, value, ox, oy, oz, events))
		}
	}()

	done := make(chan struct{})
	go func() { wg.Wait(); close(done); w.Close() }()

	// Reader: the SAME [len:u32][payload] framing runCommand.ts's splitFrames applies to
	// every dedicated stream fd, with the SAME MAX_FRAME_BYTES bound and the SAME
	// "stop on a bad length" behavior handleInteriorFd has (see this file's header
	// comment). desynced is set the first time a declared length is nonsense, mirroring
	// the ext host's `error` return from splitFrames and its "stopping stream" log line.
	var desyncErr string
	readErr := make(chan struct{})
	go func() {
		defer close(readErr)
		var hdr [4]byte
		drain := make([]byte, 65536)
		for {
			if _, err := readFullForTest(r, hdr[:]); err != nil {
				return
			}
			n := binary.LittleEndian.Uint32(hdr[:])
			if n > maxFrameBytesForTest {
				if desyncErr == "" {
					desyncErr = "bad frame length"
				}
				// The real ext host stops reading this fd entirely once desynced
				// (handleInteriorFd's deadStreams — see the header comment). This
				// test keeps draining raw bytes past that point ONLY so the still-
				// writing goroutines above don't deadlock on a full pipe buffer;
				// it does not attempt to resynchronize, matching production.
				for {
					if _, err := r.Read(drain); err != nil {
						return
					}
				}
			}
			payload := make([]byte, n)
			if _, err := readFullForTest(r, payload); err != nil {
				if desyncErr == "" {
					desyncErr = "short read after a plausible-looking length — same desync, different symptom"
				}
				return
			}
			// Only two frame shapes are ever legitimately written here: a
			// bare bead-state frame (events=nil: 4-byte tick + 4 interior rows
			// * 17 bytes + 4-byte zero event count = 76) and a one-event frame
			// (WriteEvents with one empty-Text RowEvent: 76 + one 67-byte
			// event row = 143). Any OTHER length is corruption that happened
			// to decode as a length under MAX_FRAME_BYTES — the quiet failure
			// mode the strict bound alone would miss (the TS side would still
			// desync on a LATER frame, or misdecode this one as valid garbage).
			if n != 76 && n != 143 {
				if desyncErr == "" {
					desyncErr = "frame decoded with a plausible but wrong length (mid-payload resync, not caught by MAX_FRAME_BYTES alone)"
				}
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatalf("writers did not finish in time")
	}
	w.Close()
	<-readErr

	// INVERTED from this test's original form: back when this doc's "Reproduction"
	// section was written, the deliverable was a test that FAILED (t.Fatalf on desync) —
	// appropriate for a standalone repro with no fix yet, but wrong for a PERMANENT
	// member of `go test ./...`: a test that fails on every clean run would break every
	// future `go test`/stop-checks run on this branch and everyone after it, forever,
	// for a hazard that is fully understood and intentionally reconstructed. So this
	// test now asserts the OPPOSITE polarity: it PASSES when the desync it exists to pin
	// actually occurs (desyncErr != ""), and FAILS when it does NOT — i.e. failure here
	// means the mechanism stopped reproducing (a platform/Go-runtime change in pipe
	// write behavior, not evidence that sharing a stream became safe) and this test
	// needs a fresh look, not that everything is fine.
	if desyncErr == "" {
		t.Fatalf("expected two goroutines sharing one pipe fd, each framing via two separate " +
			"Write() calls (the PRE-FIX shape of writeInteriorStreamFrame), to desync the " +
			"reader's [len:u32] framing — it did not. This test exists to PIN that mechanism " +
			"(see this file's header comment); it passing without ever desyncing means this " +
			"OS/Go-runtime's pipe write behavior no longer interleaves two writers' separate " +
			"Write() calls the way it used to, which is a platform fact worth investigating, " +
			"NOT evidence that two-Write-calls-on-a-shared-writer became safe to reintroduce.")
	}
}

// readFullForTest is a tiny io.ReadFull equivalent (avoids importing io just for this).
func readFullForTest(r *os.File, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
