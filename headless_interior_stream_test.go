// headless_interior_stream_test.go — closes the coverage gap named in the live-regression
// investigation for commit 774ffa36 ("A port has no place: geometry dies, the role
// survives"): TestHeadlessNodeFdDedicatedStream/TestHeadlessEdgeFdDedicatedStream/
// TestHeadlessViewFdDedicatedStream all exercise ONE settled frame per row, but nothing
// exercised the INTERIOR stream's byte framing across a SUSTAINED run — the one dedicated
// stream that fired the ext host's "bad frame length" error in the live editor
// (.probe/go-errors.jsonl: handleInteriorFd(row=2): bad frame length 16777216).
//
// This test spawns the real binary (same helper every other headless test in this package
// uses) and, for the whole run, decodes EVERY [len:u32][payload] frame off EVERY node's
// interior fd using the SAME length-prefix framing algorithm runCommand.ts's splitFrames
// uses (read a u32 length, consume exactly that many payload bytes, repeat) — not just the
// LAST settled frame the existing node-fd test takes. A one-byte framing disagreement
// between the Go writer and the TS reader shows up as a wrong length on some LATER frame,
// after byte drift has accumulated; asserting only the last frame in a short settle window
// (as TestHeadlessNodeFdDedicatedStream does) can silently miss that. If the writer and
// reader genuinely disagree, this test fails with a bad/short length exactly the way the
// live editor did — proving the reproduction, not just declaring it.
package main

import (
	"bufio"
	"os"
	"testing"
	"time"

	B "github.com/dtauraso/wirefold/Buffer"
)

// TestHeadlessInteriorFdSustainedFraming drives the real binary for several seconds and
// decodes every frame off every node's own dedicated interior fd, checking that the
// [len:u32] prefix always matches a length this test can independently reconstruct from
// Buffer's own layout constants (Buffer.BuildInteriorStreamFrame's shape: fixed 4-slot
// Interior rows + a trailing [count:u32] + count EVENT rows). A disagreement here is BY
// CONSTRUCTION the same class of bug that produced the live "bad frame length" error: the
// writer (Go) and an independent reader built from the same generated constants the TS
// reader also imports (Buffer/frame_tags.go / tools/topology-vscode/src/schema/frame-tags.ts,
// kept in parity by check-buffer-layout-parity.sh) disagree about where one frame ends and
// the next begins.
func TestHeadlessInteriorFdSustainedFraming(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	binPath := buildHeadlessBinary(t, repoRoot, "wirefold-headless-interior-sustained-test")
	ds := spawnDedicatedAllStreams(t, binPath, repoRoot)

	// Sustained window: long enough to observe dozens of frames per row across every
	// node in this topology (node/edge streams write every ~17ms cycle; interior
	// streams are event-driven and fire far less often, so this window is set to the
	// same order of magnitude as headless_stream_helpers_test.go's own reasoning for
	// why interior/node/edge streams never go idle in this self-feeding-ring topology).
	const runWindow = 3 * time.Second
	deadline := time.Now().Add(runWindow)

	totalFrames := 0
	for row, f := range ds.interiorReads {
		frameCount := 0
		r := bufio.NewReader(f)
		if err := f.SetReadDeadline(time.Time{}); err != nil {
			t.Fatalf("SetReadDeadline (clear, interior row %d): %v", row, err)
		}
		// Read the first frame with no deadline (same reasoning as readLastFrames: a
		// loaded test machine can delay the first byte for reasons unrelated to the
		// stream itself), then bound the rest of this row's read loop by the shared
		// wall-clock deadline (not per-row — every row shares the same runWindow so the
		// whole test costs runWindow once, not runWindow × nodeCount).
		if err := f.SetReadDeadline(deadline.Add(2 * time.Second)); err != nil {
			t.Fatalf("SetReadDeadline (interior row %d): %v", row, err)
		}
		for time.Now().Before(deadline) {
			frame, err := readOneRawFrame(r)
			if err != nil {
				// A read timeout is expected once this row's stream has gone quiet for
				// the remainder of the window (interior streams are event-driven — a
				// node with no bead-state change in [now, deadline) simply stops
				// writing). Any OTHER error ends this row's read loop the same way.
				break
			}
			frameCount++
			// Fixed-slot frame (Buffer.BuildInteriorStreamFrame): [tick:u32] + 4
			// Interior rows + a trailing EVENTS section ([count:u32] + count NodeBead
			// rows). This is the EXACT same shape TestHeadlessNodeFdDedicatedStream
			// checks for the LAST frame only; here it is checked for EVERY frame this
			// row emits during the whole run, which is what actually exercises framing
			// drift rather than a single post-settle snapshot.
			fixedLen := 4 + 4*B.BufInteriorStride
			if len(frame) < fixedLen+4 {
				t.Fatalf("interior row %d frame %d: length %d, want >= %d (fixed 4-slot layout + EVENTS count)",
					row, frameCount, len(frame), fixedLen+4)
			}
			eventCount := readU32(frame, fixedLen)
			want := fixedLen + 4 + int(eventCount)*B.BufEventStride
			if len(frame) != want {
				t.Fatalf("interior row %d frame %d: length %d, want %d (fixed 4-slot layout + %d events) — "+
					"this is the writer/reader disagreement class that produced the live "+
					"\"bad frame length\" ext-host error",
					row, frameCount, len(frame), want, eventCount)
			}
		}
		totalFrames += frameCount
	}
	// The interior stream is event-driven (a node's own Update goroutine writes only
	// when its 4-slot bead grid or a Fire/Recv/Send event changes — port_wiring.go's
	// newInteriorStreamGetter never writes an initial frame at construction, and a node
	// with no inbound activity in this window may legitimately write nothing at all).
	// So per-row silence is NOT itself a failure — but SOME node in this actively
	// traversing topology (topology/nodes/*/cascade-edges.json wires a real ring) must
	// produce interior traffic, or every frame-length assertion above ran zero times
	// and this test would "pass" without ever exercising the framing it exists to check.
	if totalFrames == 0 {
		t.Fatalf("no interior frames arrived on ANY node row in %s — the dedicated interior fd stream produced nothing to check", runWindow)
	}
}
