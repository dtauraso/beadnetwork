// stdin_input_integration_test.go — tests that drive Wiring's own stdin dispatch
// (applyEdit, RunStdinReader, newMoveDispatch) using the inputcodec package's real
// encode/decode. Pure decode/byte-layout tests live in inputcodec/input_codec_test.go —
// this file covers what only THIS package decides: applying a decoded message and
// persisting the result.

package Wiring

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// enumIndex finds s's position in list — used here to build the exact wire bytes
// inputcodec.DecodeInputRecord expects, without depending on inputcodec's own _test.go
// encode helpers (a _test.go file's symbols are not visible outside its own package).
func enumIndex(list []string, s string) byte {
	for i, v := range list {
		if v == s {
			return byte(i)
		}
	}
	return 0
}

// frameOverlaysToggleRecord builds the exact bytes an overlays-toggle edit record carries
// on the wire: [InKindEditUpdate][entityKind=overlays][InOverlayAttrToggle][flagId].
func frameOverlaysToggleRecord(flag string) []byte {
	return []byte{
		inputcodec.InKindEditUpdate,
		enumIndex(inputcodec.InUpdateKinds, "overlays"),
		inputcodec.InOverlayAttrToggle,
		enumIndex(inputcodec.InOverlayFlags, flag),
	}
}

// frameRecord wraps a record body with the [len:u32-LE] transport frame RunStdinReader
// expects.
func frameRecord(rec []byte) []byte {
	return append(binary.LittleEndian.AppendUint32(nil, uint32(len(rec))), rec...)
}

// TestSavePersistsCurrentOverlayState applies an overlays TOGGLE edit (flipping Go's held
// state), then a save, and asserts overlays.json reflects the CURRENT (post-edit) state —
// not a stale/empty snapshot. This is the "Go persists its own current topology" guarantee.
func TestSavePersistsCurrentOverlayState(t *testing.T) {
	root := t.TempDir()
	md, err := newMoveDispatch(map[string]nodeGeom{}, map[string]inputcodec.EdgeEndpoints{}, nil, nil, nil, wire.NewRealClock(), nil, 0)
	if err != nil {
		t.Fatalf("newMoveDispatch: %v", err)
	}
	// overlaysVisible defaults true; toggle flips it to false.
	toggle, ok := inputcodec.DecodeInputRecord(frameOverlaysToggleRecord("overlays"))
	if !ok {
		t.Fatal("decode toggle failed")
	}
	applyEdit(toggle, md, nil, nil)
	if err := writeSceneOverlays(scenepaths.OverlaysFilePath(root), md.ui.ov); err != nil {
		t.Fatalf("writeSceneOverlays: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "view", "overlays.json"))
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("overlays.json invalid: %v", err)
	}
	if string(obj["overlaysActive"]) != "false" {
		t.Fatalf("overlaysActive=%s want false (toggled-off state should persist)", obj["overlaysActive"])
	}
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
	md, err := newMoveDispatch(map[string]nodeGeom{}, map[string]inputcodec.EdgeEndpoints{}, nil, nil, nil, wire.NewRealClock(), nil, 0)
	if err != nil {
		t.Fatalf("newMoveDispatch: %v", err)
	}
	md.EnableEditPersist(root) // arms overlaysPersist so `save` can write overlays.json
	readerDone := make(chan struct{})
	go func() {
		RunStdinReader(ctx, pr, inputcodec.SlotRegistry{}, md, nil, nil)
		close(readerDone)
	}()
	// RunStdinReader now flushes pending debounced persisters (writes under root) on its
	// own clean-shutdown return path. Wait for that goroutine to actually finish before this
	// test's t.TempDir() cleanup removes root, or the flush can race the RemoveAll.
	defer func() { <-readerDone }()

	frame := frameRecord([]byte{inputcodec.InKindSave})
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
		RunStdinReader(ctx, pr, inputcodec.SlotRegistry{}, nil, nil, nil)
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
