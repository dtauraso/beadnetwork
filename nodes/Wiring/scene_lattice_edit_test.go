package Wiring

// scene_lattice_edit_test.go — asserts what ONE goroutine (the view-owner, driving
// applyUpdateScene directly, per docs/process/testing-shape.md) decided on a scene/latticePoints
// edit: reject anything outside 4..64 or not a multiple of 4 (never panic, never persist,
// never broadcast), accept a valid count (persist it, install it on md.UI.LatticePoints,
// broadcast it), and never block a registered-but-full LatticeIn channel.

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
)

// TestApplyUpdateSceneLatticePointsIgnoresInvalidCounts: 0, 3, 25, 65, and -4 must each be
// a no-op — md.UI.LatticePoints stays at whatever it was, and nothing is written to disk.
func TestApplyUpdateSceneLatticePointsIgnoresInvalidCounts(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)
	md.UI.LatticePoints = 24 // known starting value

	for _, bad := range []int{0, 3, 25, 65, -4} {
		msg := inputcodec.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: bad}
		applyUpdate(msg, md, nil, nil)
		if md.UI.LatticePoints != 24 {
			t.Fatalf("latticePoints=%d: md.UI.LatticePoints changed to %d, want unchanged 24", bad, md.UI.LatticePoints)
		}
		if _, found := loadSceneLattice(scenepaths.LatticeFilePath(root)); found {
			t.Fatalf("latticePoints=%d: lattice.json was written for an invalid count", bad)
		}
	}
}

// TestApplyUpdateSceneLatticePointsAcceptsValidCounts: every valid boundary/interior value
// (4, 12, 24, 64) is installed on md.UI.LatticePoints and persisted to lattice.json.
func TestApplyUpdateSceneLatticePointsAcceptsValidCounts(t *testing.T) {
	for _, good := range []int{4, 12, 24, 64} {
		root := writeTree(t)
		md := loadTreeMD(t, root)
		md.EnableEditPersist(root)

		msg := inputcodec.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: good}
		applyUpdate(msg, md, nil, nil)

		if md.UI.LatticePoints != int32(good) {
			t.Fatalf("latticePoints=%d: md.UI.LatticePoints = %d, want %d", good, md.UI.LatticePoints, good)
		}
		got, found := loadSceneLattice(scenepaths.LatticeFilePath(root))
		if !found {
			t.Fatalf("latticePoints=%d: lattice.json was not written", good)
		}
		if got != int32(good) {
			t.Fatalf("latticePoints=%d: lattice.json carries %d", good, got)
		}
	}
}

// TestBroadcastLatticePointsReachesEveryRegisteredChannel: a channel registered in
// md.latticeIns receives the broadcast count. This is what ONE goroutine (the stdin
// reader) decided to send onto directory entries it owns — not a claim about a second
// goroutine actually reading it (docs/process/testing-shape.md).
func TestBroadcastLatticePointsReachesEveryRegisteredChannel(t *testing.T) {
	md := &MoveDispatch{}
	chA := make(chan int32, 1)
	chB := make(chan int32, 1)
	md.inboxes.lattice = map[string]chan int32{"1": chA, "2": chB}

	md.BroadcastLatticePoints(12)

	for id, ch := range map[string]chan int32{"1": chA, "2": chB} {
		select {
		case got := <-ch:
			if got != 12 {
				t.Fatalf("node %s: latticeIns got %d, want 12", id, got)
			}
		default:
			t.Fatalf("node %s: BroadcastLatticePoints did not deliver onto its channel", id)
		}
	}
}

// TestBroadcastLatticePointsDoesNotBlockOnAFullChannel: a channel that already holds a
// stale pending value must be drained and replaced, never blocked on — the sender (the
// stdin-reader goroutine) must never stall because one node's own goroutine is asleep or
// mid-cycle.
func TestBroadcastLatticePointsDoesNotBlockOnAFullChannel(t *testing.T) {
	md := &MoveDispatch{}
	ch := make(chan int32, 1)
	ch <- 8 // pre-fill: simulates a stale pending value nobody has drained yet
	md.inboxes.lattice = map[string]chan int32{"1": ch}

	// Called directly, on this test's own goroutine: a non-blocking drain-then-send
	// either returns immediately (pass) or the test itself hangs and the runner times it
	// out (fail) — no second goroutine is needed to observe that.
	md.BroadcastLatticePoints(24)

	select {
	case got := <-ch:
		if got != 24 {
			t.Fatalf("channel holds %d after broadcast, want the latest value 24", got)
		}
	default:
		t.Fatalf("BroadcastLatticePoints left the channel empty")
	}
}
