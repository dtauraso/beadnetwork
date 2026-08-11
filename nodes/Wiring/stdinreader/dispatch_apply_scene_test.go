package stdinreader

// dispatch_apply_scene_test.go — asserts what ONE goroutine (the view-owner, driving
// applyUpdateScene directly, per docs/process/testing-shape.md) decided on a scene/latticePoints
// edit: reject anything outside 4..64 or not a multiple of 4 (never panic, never persist,
// never broadcast), accept a valid count (persist it, install it on md.UI.LatticePoints,
// broadcast it). Moved here from nodes/Wiring/dispatch (§30,
// docs/planning/movedispatch-decomposition.md) alongside applyUpdateScene itself; the two
// BroadcastLatticePoints-channel tests that used to sit in the same file stayed behind in
// nodes/Wiring/dispatch/scene_lattice_broadcast_test.go — they exercise
// MoveDispatch.BroadcastLatticePoints directly, not the edit-routing this file covers.

import (
	"context"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

// TestApplyUpdateSceneLatticePointsIgnoresInvalidCounts: 0, 3, 25, 65, and -4 must each be
// a no-op — md.UI.LatticePoints stays at whatever it was, and nothing is written to disk.
func TestApplyUpdateSceneLatticePointsIgnoresInvalidCounts(t *testing.T) {
	root := writeMinimalTree(t)
	md := loadMinimalMD(t, root)
	md.EnableEditPersist(root)
	md.UI.LatticePoints = 24 // known starting value

	for _, bad := range []int{0, 3, 25, 65, -4} {
		msg := inputcodec.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: bad}
		applyUpdate(context.Background(), msg, md, nil, nil)
		if md.UI.LatticePoints != 24 {
			t.Fatalf("latticePoints=%d: md.UI.LatticePoints changed to %d, want unchanged 24", bad, md.UI.LatticePoints)
		}
		if _, found := scenepersist.LoadSceneLattice(scenepaths.LatticeFilePath(root)); found {
			t.Fatalf("latticePoints=%d: lattice.json was written for an invalid count", bad)
		}
	}
}

// TestApplyUpdateSceneLatticePointsAcceptsValidCounts: every valid boundary/interior value
// (4, 12, 24, 64) is installed on md.UI.LatticePoints and persisted to lattice.json.
func TestApplyUpdateSceneLatticePointsAcceptsValidCounts(t *testing.T) {
	for _, good := range []int{4, 12, 24, 64} {
		root := writeMinimalTree(t)
		md := loadMinimalMD(t, root)
		md.EnableEditPersist(root)

		msg := inputcodec.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: good}
		applyUpdate(context.Background(), msg, md, nil, nil)

		if md.UI.LatticePoints != int32(good) {
			t.Fatalf("latticePoints=%d: md.UI.LatticePoints = %d, want %d", good, md.UI.LatticePoints, good)
		}
		got, found := scenepersist.LoadSceneLattice(scenepaths.LatticeFilePath(root))
		if !found {
			t.Fatalf("latticePoints=%d: lattice.json was not written", good)
		}
		if got != int32(good) {
			t.Fatalf("latticePoints=%d: lattice.json carries %d", good, got)
		}
	}
}
