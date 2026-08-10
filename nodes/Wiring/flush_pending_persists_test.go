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

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
)

// TestQuantOffsetScheduleWritesSynchronously proves persistQuantOffset writes a position
// update to disk immediately, with no timer/flush step required.
func TestQuantOffsetScheduleWritesSynchronously(t *testing.T) {
	root := writeTree(t)
	nm := &nodeGeometry{id: "1", persistRoot: root}

	newScene := geom.Polar{R: 55.5, Theta: 0.4, Phi: -1.1}
	nm.persistQuantOffset(quantoffset.QuantizedOffset{ITheta: 3, IPhi: 4, IR: 5}, newScene)

	raw, err := os.ReadFile(positionfile.FilePath(root, "1"))
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

	nm, ok := md.mr.nodeGeoms["1"]
	if !ok {
		t.Fatal("no nodeMover for src")
	}
	newScene := geom.Polar{R: 61.0, Theta: 0.2, Phi: 0.9}
	nm.persistQuantOffset(quantoffset.QuantizedOffset{ITheta: 1, IPhi: 2, IR: 3}, newScene)

	raw, err := os.ReadFile(positionfile.FilePath(root, "1"))
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
	nm := &nodeGeometry{id: "x"}                                       // persistRoot == "" — unarmed
	nm.persistQuantOffset(quantoffset.QuantizedOffset{}, geom.Polar{}) // must not panic
}
