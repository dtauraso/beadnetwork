package Node1

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// The straightening loop's whole reactive rule now lives on THIS node kind's own
// goroutine (stepTilt) — this asserts what that ONE goroutine's own
// method decides, per docs/testing-shape.md: index arithmetic only, no mover/second
// goroutine involved.

// Node1 always SUBTRACTS one step of π/12, whichever side of perpendicular it is on — the
// direction is a property of the KIND, not of where the tilt currently sits. Its partner
// moves the opposite way by the same step, so a pair turns symmetrically.
func TestStepAlwaysMovesTheSameDirection(t *testing.T) {
	below := &Node{TiltThetaIdx: 3}
	if moved := below.stepTilt(); !moved || below.TiltThetaIdx != 2 {
		t.Fatalf("from below perpendicular, want moved=true thetaIdx=2, got moved=%v thetaIdx=%d", moved, below.TiltThetaIdx)
	}

	above := &Node{TiltThetaIdx: 9}
	if moved := above.stepTilt(); !moved || above.TiltThetaIdx != 8 {
		t.Fatalf("from above perpendicular, want moved=true thetaIdx=8, got moved=%v thetaIdx=%d", moved, above.TiltThetaIdx)
	}
}

// AT perpendicular, a call changes nothing and reports no move — this is the loop's
// termination, not a missed case.
func TestStepStopsAtPerpendicular(t *testing.T) {
	n := &Node{TiltThetaIdx: Wiring.PerpendicularThetaIdx}
	if moved := n.stepTilt(); moved {
		t.Fatalf("at perpendicular, want moved=false, got true (index now %d)", n.TiltThetaIdx)
	}
	if n.TiltThetaIdx != Wiring.PerpendicularThetaIdx {
		t.Fatalf("at perpendicular, index must not move; got %d, want %d", n.TiltThetaIdx, Wiring.PerpendicularThetaIdx)
	}
}

// applyTiltEdit is what the RESET button (TiltResetButton.tsx) and the tilt-angle panel
// both drive; this asserts THIS node kind's own decision for a reset, from any starting
// indices: both indices land on 0 (the start position), and — unlike an adjust — no bead
// is placed ("the kick" only fires for a panel nudge, never a stop-and-return).
func TestApplyTiltEditResetReturnsBothIndicesToZero(t *testing.T) {
	for _, start := range []struct{ theta, phi int32 }{
		{0, 0},
		{5, -3},
		{-9, 12},
		{Wiring.PerpendicularThetaIdx, 4},
	} {
		n := &Node{TiltThetaIdx: start.theta, TiltPhiIdx: start.phi}
		placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
		if placeBead {
			t.Fatalf("reset from theta=%d phi=%d must place no bead, got placeBead=true", start.theta, start.phi)
		}
		if n.TiltThetaIdx != 0 || n.TiltPhiIdx != 0 {
			t.Fatalf("reset from theta=%d phi=%d: want both indices 0, got theta=%d phi=%d", start.theta, start.phi, n.TiltThetaIdx, n.TiltPhiIdx)
		}
	}
}

// The drawn coplanar normal must sit exactly +6 steps (90°) in θ from the tilt, φ
// unchanged, for Node1 — pure index arithmetic reusing Wiring.PerpendicularThetaIdx, not
// a cross product (Part 1 of this task).
func TestCoplanarNormalIsPlusSixStepsInTheta(t *testing.T) {
	for _, start := range []struct{ theta, phi int32 }{
		{0, 0}, {5, -3}, {-9, 12},
	} {
		n := &Node{TiltThetaIdx: start.theta, TiltPhiIdx: start.phi}
		norm := n.coplanarNormal()
		if norm.ThetaIdx != start.theta+Wiring.PerpendicularThetaIdx {
			t.Fatalf("coplanarNormal theta: want tilt+%d=%d, got %d", Wiring.PerpendicularThetaIdx, start.theta+Wiring.PerpendicularThetaIdx, norm.ThetaIdx)
		}
		if norm.PhiIdx != start.phi {
			t.Fatalf("coplanarNormal phi must equal tilt phi unchanged: want %d, got %d", start.phi, norm.PhiIdx)
		}
	}
}

// RESET must also drain any value already sitting on VectorIn (depth-1 latest-wins), so
// it cannot arrive on the NEXT cycle and immediately step the tilt again, undoing the
// reset. Verified by observing the channel is empty afterward.
func TestApplyTiltEditResetDrainsVectorIn(t *testing.T) {
	vectorIn := make(chan Wiring.TiltVectorMsg, 1)
	vectorIn <- Wiring.TiltVectorMsg{ThetaIdx: 99, PhiIdx: -1}
	n := &Node{TiltThetaIdx: 5, TiltPhiIdx: -2, VectorIn: vectorIn}
	n.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
	select {
	case v := <-vectorIn:
		t.Fatalf("VectorIn must be drained by reset; still holds %+v", v)
	default:
	}
}

// A non-reset edit (the panel's ±1 click) still unconditionally applies its delta and
// reports a bead should be placed — the RESET addition must not change this existing path.
func TestApplyTiltEditAdjustStillPlacesBead(t *testing.T) {
	n := &Node{TiltThetaIdx: 3, TiltPhiIdx: 1}
	placeBead := n.applyTiltEdit(Wiring.TiltEditMsg{Axis: "theta", Up: true})
	if !placeBead {
		t.Fatalf("adjust must place a bead, got placeBead=false")
	}
	if n.TiltThetaIdx != 4 || n.TiltPhiIdx != 1 {
		t.Fatalf("adjust theta up: want theta=4 phi=1, got theta=%d phi=%d", n.TiltThetaIdx, n.TiltPhiIdx)
	}
}

// outgoingVector reverses the coplanar normal by 180° in θ — Node1's own direction is
// −12 index steps (−180°, two of the 6-step quarter turns) — and leaves φ untouched.
// This is THIS node's own arithmetic, asserted without any channel involved.
func TestOutgoingVectorIsMinus180StepsInThetaOnly(t *testing.T) {
	n := &Node{TiltThetaIdx: 3, TiltPhiIdx: -7}
	norm := n.coplanarNormal()
	if norm.ThetaIdx != 3+Wiring.PerpendicularThetaIdx || norm.PhiIdx != -7 {
		t.Fatalf("coplanarNormal: want theta=%d phi=-7, got theta=%d phi=%d",
			3+Wiring.PerpendicularThetaIdx, norm.ThetaIdx, norm.PhiIdx)
	}
	out := n.outgoingVector()
	wantTheta := norm.ThetaIdx - 2*Wiring.PerpendicularThetaIdx
	if out.ThetaIdx != wantTheta {
		t.Fatalf("outgoingVector theta: want %d (norm - 12 steps), got %d", wantTheta, out.ThetaIdx)
	}
	if out.PhiIdx != norm.PhiIdx {
		t.Fatalf("outgoingVector must leave phi unchanged: norm phi=%d, out phi=%d", norm.PhiIdx, out.PhiIdx)
	}
}

// stepTowardPerpendicularFromVector is Node1's own step decision on a vector-channel
// arrival: away from perpendicular it moves one π/12 step (same direction stepTilt
// always takes for this kind), and reports the move.
func TestStepFromVectorMovesOneStepAwayFromPerpendicular(t *testing.T) {
	n := &Node{TiltThetaIdx: 3}
	received := Wiring.TiltVectorMsg{ThetaIdx: 99, PhiIdx: -5}
	if moved := n.stepTowardPerpendicularFromVector(received); !moved || n.TiltThetaIdx != 2 {
		t.Fatalf("want moved=true thetaIdx=2, got moved=%v thetaIdx=%d", moved, n.TiltThetaIdx)
	}
}

// AT perpendicular, an arriving vector still steps nothing and the caller must send
// nothing — this is the exchange's stop condition, asserted at the decision method
// itself, not by observing two goroutines communicate.
func TestStepFromVectorStopsAtPerpendicular(t *testing.T) {
	n := &Node{TiltThetaIdx: Wiring.PerpendicularThetaIdx}
	received := Wiring.TiltVectorMsg{ThetaIdx: 1, PhiIdx: 1}
	if moved := n.stepTowardPerpendicularFromVector(received); moved {
		t.Fatalf("at perpendicular, want moved=false, got true (index now %d)", n.TiltThetaIdx)
	}
	if n.TiltThetaIdx != Wiring.PerpendicularThetaIdx {
		t.Fatalf("at perpendicular, index must not move; got %d, want %d", n.TiltThetaIdx, Wiring.PerpendicularThetaIdx)
	}
}

// A RESET empties BOTH directions — but by each node draining the one receive end it owns,
// not by either of them evicting the other's queue. A send-only end cannot drain itself (in
// Go only a receiver empties a channel), so this asserts the pair-level property: reset both
// nodes and nothing is left anywhere.
func TestResettingBothNodesEmptiesBothDirections(t *testing.T) {
	oneToTwo := make(chan Wiring.TiltVectorMsg, 1)
	twoToOne := make(chan Wiring.TiltVectorMsg, 1)
	// Node1 sends on oneToTwo and receives on twoToOne; its partner is the mirror image.
	one := &Node{TiltThetaIdx: 4, TiltPhiIdx: 2, VectorOut: oneToTwo, VectorIn: twoToOne}
	partnerIn := oneToTwo // what the other node owns the receive end of

	// A stale direction is in flight BOTH ways when reset is pressed.
	oneToTwo <- Wiring.TiltVectorMsg{ThetaIdx: 9}
	twoToOne <- Wiring.TiltVectorMsg{ThetaIdx: 9}

	one.applyTiltEdit(Wiring.TiltEditMsg{Reset: true})
	if one.TiltThetaIdx != 0 || one.TiltPhiIdx != 0 {
		t.Fatalf("reset must zero both indices; got theta=%d phi=%d", one.TiltThetaIdx, one.TiltPhiIdx)
	}
	if _, ok := Wiring.PollRecvVector(twoToOne); ok {
		t.Fatal("reset left a value on the end this node owns")
	}
	// The other direction is still full — this node cannot clear it, and that is the point:
	// the partner's own reset is what empties it, which the button sends too.
	if _, ok := Wiring.PollRecvVector(partnerIn); !ok {
		t.Fatal("expected the outward direction to still hold its stale value before the partner resets")
	}
	if _, ok := Wiring.PollRecvVector(partnerIn); ok {
		t.Fatal("after the partner's own drain, both directions must be empty")
	}
}

// A received reset zeroes this node and REPLIES WITH NOTHING — a reply would bounce the
// reset back and forth forever instead of ending the exchange.
func TestReceivedResetZeroesAndDoesNotReply(t *testing.T) {
	out := make(chan Wiring.TiltVectorMsg, 1)
	in := make(chan Wiring.TiltVectorMsg, 1)
	n := &Node{TiltThetaIdx: 5, TiltPhiIdx: 3, VectorOut: out, VectorIn: in}

	in <- Wiring.TiltVectorMsg{Reset: true}
	n.handleVectorCycle()

	if n.TiltThetaIdx != 0 || n.TiltPhiIdx != 0 {
		t.Fatalf("a received reset must zero both indices; got theta=%d phi=%d", n.TiltThetaIdx, n.TiltPhiIdx)
	}
	if v, ok := Wiring.PollRecvVector(out); ok {
		t.Fatalf("a received reset must not be replied to; got %+v", v)
	}
}
