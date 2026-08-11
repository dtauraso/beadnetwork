package dispatch_test

// flush_pending_persists_test.go — the load-path half of the quant-offset persistence
// coverage; the bare-NodeGeometry half moved to
// nodes/Wiring/nodeactor/quant_offset_persist_test.go in
// docs/planning/movedispatch-decomposition.md §20 (persistQuantOffset is unexported and
// this file no longer constructs a bare *nodeactor.NodeGeometry, so it stayed here only
// for the case that genuinely needs the real loader/tree fixtures).
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
)

// TestMoveDispatchQuantOffsetScheduleWritesThroughEnableEditPersist exercises the
// node-owned write path (EnableEditPersist sets nm.persistRoot on every mover) and
// confirms it reaches disk. Drives CommitQuantOffset (the exported commit-path method,
// node_geometry_accessors.go) rather than the unexported persistQuantOffset directly —
// CommitQuantOffset persists its committedPolar argument's R/Theta/Phi VERBATIM (see that
// method's own doc comment: persistQuantOffset writes `scene` as given, never a value
// derived from the measured offset), so asserting scenePolarR against newScene.R proves
// exactly the same "reaches disk through the real EnableEditPersist wiring" fact the
// former direct persistQuantOffset(literal-offset, newScene) call proved.
func TestMoveDispatchQuantOffsetScheduleWritesThroughEnableEditPersist(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)

	nm, ok := md.MR.NodeGeoms()["1"]
	if !ok {
		t.Fatal("no nodeMover for src")
	}
	newScene := geom.Polar{R: 61.0, Theta: 0.2, Phi: 0.9}
	nm.CommitQuantOffset(newScene)

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
