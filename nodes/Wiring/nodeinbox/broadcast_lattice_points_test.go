// broadcast_lattice_points_test.go — asserts what BroadcastLatticePoints itself decided:
// reach every registered channel, and never block on a channel already holding a stale
// pending value (docs/process/testing-shape.md). Moved from nodes/Wiring/dispatch's
// scene_lattice_broadcast_test.go (docs/planning/movedispatch-decomposition.md, the
// remainder cluster) alongside the deletion of MoveDispatch.BroadcastLatticePoints, a pure
// single-owner forward onto *NodeInboxes — every caller now addresses md.Inboxes directly.
package nodeinbox

import "testing"

// TestBroadcastLatticePointsReachesEveryRegisteredChannel: a channel registered via
// ClaimLatticeIn receives the broadcast count. This is what ONE goroutine (the stdin
// reader) decided to send onto directory entries it owns — not a claim about a second
// goroutine actually reading it (docs/process/testing-shape.md).
func TestBroadcastLatticePointsReachesEveryRegisteredChannel(t *testing.T) {
	var ib NodeInboxes
	chA := make(chan int32, 1)
	chB := make(chan int32, 1)
	ib.ClaimLatticeIn("1", chA)
	ib.ClaimLatticeIn("2", chB)

	ib.BroadcastLatticePoints(12)

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
	var ib NodeInboxes
	ch := make(chan int32, 1)
	ch <- 8 // pre-fill: simulates a stale pending value nobody has drained yet
	ib.ClaimLatticeIn("1", ch)

	// Called directly, on this test's own goroutine: a non-blocking drain-then-send
	// either returns immediately (pass) or the test itself hangs and the runner times it
	// out (fail) — no second goroutine is needed to observe that.
	ib.BroadcastLatticePoints(24)

	select {
	case got := <-ch:
		if got != 24 {
			t.Fatalf("channel holds %d after broadcast, want the latest value 24", got)
		}
	default:
		t.Fatalf("BroadcastLatticePoints left the channel empty")
	}
}
