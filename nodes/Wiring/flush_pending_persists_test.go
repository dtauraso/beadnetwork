package Wiring

// flush_pending_persists_test.go — pins that a quant-offset persist write lands on disk
// SYNCHRONOUSLY, with no clean-shutdown flush needed to guard against loss.
//
// This used to pin MoveDispatch.flushPendingPersists, a `defer` in RunStdinReader that
// flushed every persister's pending debounced value on process exit — without it, a drag
// landing within the 250ms debounce window of exit was silently abandoned. That whole
// mechanism (debouncedPersister, flushPending, flushPendingPersists) was removed: each
// persister now writes the moment its value changes, so there is nothing left pending at
// exit to lose. See scene_persist.go's header comment for the reasoning.

import (
	"encoding/json"
	"os"
	"testing"
)

// TestQuantOffsetScheduleWritesSynchronously proves persistQuantOffset writes a position
// update to disk immediately, with no timer/flush step required.
func TestQuantOffsetScheduleWritesSynchronously(t *testing.T) {
	root := writeTree(t)
	nm := &nodeMover{id: "src", persistRoot: root}

	newScene := polar{R: 55.5, Theta: 0.4, Phi: -1.1}
	nm.persistQuantOffset(quantizedOffset{iTheta: 3, iPhi: 4, iR: 5}, newScene)

	raw, err := os.ReadFile(positionFilePath(root, "src"))
	if err != nil {
		t.Fatalf("read position.json: %v", err)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal position.json: %v", err)
	}
	var gotR, gotTheta, gotPhi float64
	if err := json.Unmarshal(obj["scenePolarR"], &gotR); err != nil {
		t.Fatalf("scenePolarR: %v", err)
	}
	if err := json.Unmarshal(obj["scenePolarTheta"], &gotTheta); err != nil {
		t.Fatalf("scenePolarTheta: %v", err)
	}
	if err := json.Unmarshal(obj["scenePolarPhi"], &gotPhi); err != nil {
		t.Fatalf("scenePolarPhi: %v", err)
	}
	if gotR != newScene.R || gotTheta != newScene.Theta || gotPhi != newScene.Phi {
		t.Fatalf("schedule did not synchronously persist the drag: got (%v,%v,%v) want (%v,%v,%v)",
			gotR, gotTheta, gotPhi, newScene.R, newScene.Theta, newScene.Phi)
	}
}

// TestMoveDispatchQuantOffsetScheduleWritesThroughEnableEditPersist exercises the
// node-owned write path (EnableEditPersist sets nm.persistRoot on every mover) and
// confirms it reaches disk.
func TestMoveDispatchQuantOffsetScheduleWritesThroughEnableEditPersist(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)

	nm, ok := md.mr.nodeMovers["src"]
	if !ok {
		t.Fatal("no nodeMover for src")
	}
	newScene := polar{R: 61.0, Theta: 0.2, Phi: 0.9}
	nm.persistQuantOffset(quantizedOffset{iTheta: 1, iPhi: 2, iR: 3}, newScene)

	raw, err := os.ReadFile(positionFilePath(root, "src"))
	if err != nil {
		t.Fatalf("read position.json: %v", err)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal position.json: %v", err)
	}
	var gotR float64
	if err := json.Unmarshal(obj["scenePolarR"], &gotR); err != nil {
		t.Fatalf("scenePolarR: %v", err)
	}
	if gotR != newScene.R {
		t.Fatalf("quant-offset schedule did not write: got scenePolarR=%v want %v", gotR, newScene.R)
	}
}

// TestQuantOffsetScheduleNilSafe confirms an unarmed (empty persistRoot) nodeMover does not
// panic on a persist call — tests/headless contexts construct a MoveDispatch without
// EnableEditPersist.
func TestQuantOffsetScheduleNilSafe(t *testing.T) {
	nm := &nodeMover{id: "x"}                         // persistRoot == "" — unarmed
	nm.persistQuantOffset(quantizedOffset{}, polar{}) // must not panic
	nm.persistLocalPolars(nil, dir{})                 // must not panic
}
