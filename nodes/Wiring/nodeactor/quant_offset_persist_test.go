// quant_offset_persist_test.go — pins that a quant-offset persist write lands on disk
// SYNCHRONOUSLY, with no clean-shutdown flush needed to guard against loss. Moved here
// from package Wiring's flush_pending_persists_test.go in
// docs/planning/movedispatch-decomposition.md §20: these two cases drive persistQuantOffset
// directly on a bare NodeGeometry, with no MoveDispatch/loader involved, so they are a
// single-goroutine subject entirely within this package (docs/process/testing-shape.md's
// persistence exception). The third case in that file (EnableEditPersist's real load path)
// stayed in package Wiring, since it needs the real loader/tree fixtures.
//
// This used to pin MoveDispatch.flushPendingPersists, a `defer` in RunStdinReader that
// flushed every persister's pending debounced value on process exit — without it, a drag
// landing within the 250ms debounce window of exit was silently abandoned. That whole
// mechanism (debouncedPersister, flushPending, flushPendingPersists) was removed: each
// persister now writes the moment its value changes, so there is nothing left pending at
// exit to lose. See package Wiring's scene_persist.go header comment for the reasoning.

package nodeactor

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
	root := t.TempDir()
	nm := &NodeGeometry{id: "1", persistRoot: root}

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

// TestQuantOffsetScheduleNilSafe confirms an unarmed (empty persistRoot) NodeGeometry does
// not panic on a persist call — tests/headless contexts construct a MoveDispatch without
// EnableEditPersist.
func TestQuantOffsetScheduleNilSafe(t *testing.T) {
	nm := &NodeGeometry{id: "x"}                                       // persistRoot == "" — unarmed
	nm.persistQuantOffset(quantoffset.QuantizedOffset{}, geom.Polar{}) // must not panic
}
