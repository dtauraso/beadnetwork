package Node2

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// The straightening loop's whole reactive rule now lives on THIS node kind's own
// goroutine (stepTilt) — this asserts what that ONE goroutine's own
// method decides, per docs/testing-shape.md: index arithmetic only, no mover/second
// goroutine involved.

// Node2 always ADDS one step of π/12, whichever side of perpendicular it is on — the
// direction is a property of the KIND, not of where the tilt currently sits. Its partner
// moves the opposite way by the same step, so a pair turns symmetrically.
func TestStepAlwaysMovesTheSameDirection(t *testing.T) {
	below := &Node{TiltThetaIdx: 3}
	if moved := below.stepTilt(); !moved || below.TiltThetaIdx != 4 {
		t.Fatalf("from below perpendicular, want moved=true thetaIdx=4, got moved=%v thetaIdx=%d", moved, below.TiltThetaIdx)
	}

	above := &Node{TiltThetaIdx: 9}
	if moved := above.stepTilt(); !moved || above.TiltThetaIdx != 10 {
		t.Fatalf("from above perpendicular, want moved=true thetaIdx=10, got moved=%v thetaIdx=%d", moved, above.TiltThetaIdx)
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

// outgoingVector reverses the coplanar normal by 180° in θ — Node2's own direction is
// +12 index steps (+180°, two of the 6-step quarter turns) — and leaves φ untouched.
// This is THIS node's own arithmetic, asserted without any channel involved.
func TestOutgoingVectorIsPlus180StepsInThetaOnly(t *testing.T) {
	n := &Node{TiltThetaIdx: 3, TiltPhiIdx: -7}
	norm := n.coplanarNormal()
	if norm.ThetaIdx != 3+Wiring.PerpendicularThetaIdx || norm.PhiIdx != -7 {
		t.Fatalf("coplanarNormal: want theta=%d phi=-7, got theta=%d phi=%d",
			3+Wiring.PerpendicularThetaIdx, norm.ThetaIdx, norm.PhiIdx)
	}
	out := n.outgoingVector()
	wantTheta := norm.ThetaIdx + 2*Wiring.PerpendicularThetaIdx
	if out.ThetaIdx != wantTheta {
		t.Fatalf("outgoingVector theta: want %d (norm + 12 steps), got %d", wantTheta, out.ThetaIdx)
	}
	if out.PhiIdx != norm.PhiIdx {
		t.Fatalf("outgoingVector must leave phi unchanged: norm phi=%d, out phi=%d", norm.PhiIdx, out.PhiIdx)
	}
}

// stepTowardPerpendicularFromVector is Node2's own step decision on a vector-channel
// arrival: away from perpendicular it moves one π/12 step (same direction stepTilt
// always takes for this kind), and reports the move.
func TestStepFromVectorMovesOneStepAwayFromPerpendicular(t *testing.T) {
	n := &Node{TiltThetaIdx: 3}
	received := Wiring.TiltVectorMsg{ThetaIdx: 99, PhiIdx: -5}
	if moved := n.stepTowardPerpendicularFromVector(received); !moved || n.TiltThetaIdx != 4 {
		t.Fatalf("want moved=true thetaIdx=4, got moved=%v thetaIdx=%d", moved, n.TiltThetaIdx)
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
