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
