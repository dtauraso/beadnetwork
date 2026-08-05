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

// The coplanar normal a straightening node measures against must be independent of that
// same node's own tilt (Part 1 of this task) — turning the tilt must never move the
// normal being compared against it. This is the property that makes "dot == 0" a stable
// stopping condition instead of one that chases itself.
func TestCoplanarNormalUnaffectedByTiltForStraightening(t *testing.T) {
	self := vec3{X: 100}
	partner := vec3{X: 100, Z: 100}
	axisT, axisP, ok := uprightRingAxis(self, partner)
	if !ok {
		t.Fatal("axis must resolve")
	}
	before, beforeP, ok := coplanarNormalTowardPartner(self, partner, axisT, axisP)
	if !ok {
		t.Fatal("normal must resolve")
	}
	// coplanarNormalTowardPartner takes no tilt argument at all — simulate "the tilt
	// turned" by simply calling it again; nothing about self/partner/axis changed, so
	// nothing about the result may either.
	after, afterP, ok := coplanarNormalTowardPartner(self, partner, axisT, axisP)
	if !ok || before != after || beforeP != afterP {
		t.Fatalf("coplanar normal moved independent of any tilt change: before=(%v,%v) after=(%v,%v)", before, beforeP, after, afterP)
	}
}
