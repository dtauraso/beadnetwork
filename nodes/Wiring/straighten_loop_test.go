package Wiring

import "testing"

// The straightening loop's rule now lives on the NODE KIND's own goroutine (Node1/Node2,
// nodes/Node1/node_test.go, nodes/Node2/node_test.go — one goroutine, no mover involved,
// per docs/testing-shape.md). What remains testable here, at the mover/geometry layer, is
// the shared constant the rule compares against and the geometric fact the rule's
// shortcut depends on: the coplanar normal must be independent of the tilt it's compared
// against.

// PerpendicularThetaIdx must be exactly 6 steps of CurveParamTiltVectorAngleStep from
// world +y (θ=0) — that is, π/2, the definition the whole rule rests on.
func TestPerpendicularThetaIdxIsSixSteps(t *testing.T) {
	const wantSteps = 6
	if PerpendicularThetaIdx != wantSteps {
		t.Fatalf("PerpendicularThetaIdx = %d, want %d (π/2 at %v-radian steps)", PerpendicularThetaIdx, wantSteps, CurveParamTiltVectorAngleStep)
	}
}

// The old "coplanar normal independent of tilt" test was removed along with
// coplanarNormalTowardPartner: the DRAWN normal is now defined AS a fixed ±90° offset
// FROM the tilt (nodes/Node1/node.go, nodes/Node2/node.go's coplanarNormal), so it moves
// WITH the tilt on purpose. The straightening rule's own stop condition never measured the
// drawn normal anyway — it compares TiltThetaIdx directly against PerpendicularThetaIdx
// (stepTilt, both packages), which is unaffected by this change.
