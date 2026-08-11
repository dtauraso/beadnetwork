// stdin_reader_framing_test.go — tests that drive RunStdinReader itself (framed
// partial reads, ctx-cancel shutdown) through a real pipe. Moved from
// nodes/Wiring/stdinreader/stdin_reader_integration_test.go: stdinreader no longer
// imports Wiring (it takes its three dispatch operations as function values,
// stdinreader.Handlers), so this now runs as an in-package Wiring test that imports
// stdinreader directly, calling the unexported newMoveDispatch instead of the
// now-deleted NewMoveDispatchForTest hatch.
package dispatch

import (
	"context"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/stdinreader"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// frameRecord2 wraps a record body with the [len:u32-LE] transport frame RunStdinReader
// expects. (Suffixed _2 to avoid colliding with any identically-named helper elsewhere in
// package Wiring's test files.)
func frameRecord2(rec []byte) []byte {
	return append(binary.LittleEndian.AppendUint32(nil, uint32(len(rec))), rec...)
}

// TestFramedPartialReads feeds a framed record ONE BYTE AT A TIME through a pipe and
// asserts the reader reassembles the frame and applies its side effect (a save writes
// overlays.json).
func TestFramedPartialReads(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pr, pw := io.Pipe()
	// A real (empty) dispatch so the `save` command has an overlay snapshot to persist.
	md, err := newMoveDispatch(map[string]nodegeom.NodeGeom{}, map[string]inputcodec.EdgeEndpoints{}, nil, nil, nil, clock.NewRealClock(), nil, 0)
	if err != nil {
		t.Fatalf("newMoveDispatch: %v", err)
	}
	md.EnableEditPersist(root) // arms overlaysPersist so `save` can write overlays.json
	h := stdinreader.Handlers{
		ApplyEdit:      func(msg inputcodec.StdinMsg) { ApplyEdit(msg, md, nil, nil) },
		HandleRawInput: func(msg inputcodec.StdinMsg) { HandleRawInputMsg(msg, inputcodec.SlotRegistry{}, md, nil) },
		HandleSave:     func() { HandleSaveMsg(md) },
	}
	readerDone := make(chan struct{})
	go func() {
		stdinreader.RunStdinReader(ctx, pr, h)
		close(readerDone)
	}()
	// RunStdinReader now flushes pending debounced persisters (writes under root) on its
	// own clean-shutdown return path. Wait for that goroutine to actually finish before this
	// test's t.TempDir() cleanup removes root, or the flush can race the RemoveAll.
	defer func() { <-readerDone }()

	frame := frameRecord2([]byte{inputcodec.InKindSave})
	go func() {
		for _, b := range frame {
			pw.Write([]byte{b})
			time.Sleep(100 * time.Microsecond)
		}
	}()
	overlaysPath := filepath.Join(root, "view", "overlays.json")
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(overlaysPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("partial-read frame never dispatched (overlays.json not written)")
		}
		time.Sleep(2 * time.Millisecond)
	}
	pw.Close()
}

// TestStdinReaderCancelWithoutEOF asserts the background frame-reader goroutine unwinds
// on ctx-cancel even when the pipe write end stays open (no EOF/close from the writer
// side). Before the close-on-cancel fix, the reader goroutine stayed parked in
// io.ReadFull forever once RunStdinReader itself returned via <-done — a goroutine leak
// for any in-process caller that cancels without closing r (production relies on process
// exit to reclaim it, so this went unnoticed there).
func TestStdinReaderCancelWithoutEOF(t *testing.T) {
	// Let any goroutines from earlier tests/GC settle so the baseline below is stable.
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
	base := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	pr, pw := io.Pipe()
	defer pw.Close()

	readerDone := make(chan struct{})
	go func() {
		stdinreader.RunStdinReader(ctx, pr, stdinreader.Handlers{})
		close(readerDone)
	}()

	cancel()

	// RunStdinReader's own select loop exits promptly on cancel; wait for that first so
	// this test's timing budget below is spent solely on the background frame-reader
	// goroutine's unwind, which is what the fix targets.
	select {
	case <-readerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("RunStdinReader did not return on ctx cancel")
	}

	// The background frame-reader goroutine has no direct completion signal exposed by
	// RunStdinReader, so drive it out indirectly: it exits by unblocking io.ReadFull once
	// r is closed. Detect completion via runtime.NumGoroutine returning to (at most) the
	// pre-test baseline, bounded by a short deadline — this times out before the fix
	// (goroutine stays parked forever) and passes after.
	deadline := time.Now().Add(500 * time.Millisecond)
	settled := false
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= base {
			settled = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !settled {
		t.Fatalf("background frame-reader goroutine still parked after ctx cancel (goroutine leak); NumGoroutine=%d base=%d", runtime.NumGoroutine(), base)
	}
}
