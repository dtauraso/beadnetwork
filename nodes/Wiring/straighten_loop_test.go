package Wiring

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

// The straightening loop's rule now lives on the NODE KIND's own goroutine (PairNode,
// nodes/PairNode/node_test.go — one goroutine, no mover involved,
// per docs/process/testing-shape.md). What remains testable here, at the mover/geometry layer, is
// the shared constant the rule compares against and the geometric fact the rule's
// shortcut depends on: the coplanar normal must be independent of the tilt it's compared
// against.

// PerpendicularThetaIdx must be exactly 6 steps of CurveParamTiltVectorAngleStep from
// world +y (θ=0) — that is, π/2, the definition the whole rule rests on.
func TestPerpendicularThetaIdxIsSixSteps(t *testing.T) {
	const wantSteps = 6
	if PerpendicularThetaIdx != wantSteps {
		t.Fatalf("PerpendicularThetaIdx = %d, want %d (π/2 at %v-radian steps)", PerpendicularThetaIdx, wantSteps, nodegeom.CurveParamTiltVectorAngleStep)
	}
}

// The old "coplanar normal independent of tilt" test was removed along with
// coplanarNormalTowardPartner: the DRAWN normal is now defined AS a fixed +90° offset
// FROM the tilt (nodes/PairNode/node.go's coplanarNormal, run unmodified by both nodes of a
// pair), so it moves WITH the tilt on purpose. The straightening rule's own stop condition
// never measured the drawn normal anyway — it compares TopTiltThetaIdx directly against
// PerpendicularThetaIdx (stepTilt), which is unaffected by this change.
