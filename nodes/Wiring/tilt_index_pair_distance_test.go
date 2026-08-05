package Wiring

import (
	"math"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// TestTiltIndexDistanceRoundTripsThroughEdgeStepCount pins the pair-tab model
// (task/tilt-sets-pair-distance): the distance repositionForTiltIndex sets from a tilt
// index —
//
//	D = (abs(idx) + nodeTorusSteps(srcKind) + nodeTorusSteps(dstKind)) * wire.BeadStepR
//
// must run back through edgeStepCount (chain_beads.go, never modified by this feature)
// and yield exactly abs(idx) — that round trip is the whole point of the two torus-step
// terms. Pure function of the two kind names and an index, no goroutine involved
// (docs/testing-shape.md): this is what ONE node's own arithmetic decides, not a
// cross-goroutine delivery fact.
func TestTiltIndexDistanceRoundTripsThroughEdgeStepCount(t *testing.T) {
	cases := []struct {
		name             string
		idx              int32
		srcKind, dstKind string
		wantStepCount    int
	}{
		{"zero index clamps to the documented minimum of 1", 0, "Node1", "Node2", 1},
		{"positive index", 3, "Node1", "Node2", 3},
		{"negative index — magnitude is what both sides agree on", -5, "Node1", "Node2", 5},
		{"larger magnitude, mirrored kinds", 12, "Node2", "Node1", 12},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			idx := c.idx
			if idx < 0 {
				idx = -idx
			}
			steps := int(idx) + nodeTorusSteps(c.srcKind) + nodeTorusSteps(c.dstKind)
			d := float64(steps) * wire.BeadStepR

			got := edgeStepCount(d, c.srcKind, c.dstKind)
			if got != c.wantStepCount {
				t.Fatalf("edgeStepCount(D, %q, %q) = %d, want %d (D=%.4f built from idx=%d)",
					c.srcKind, c.dstKind, got, c.wantStepCount, d, c.idx)
			}
		})
	}
}

// TestSolveTiltIndexROffsetDistanceOnly pins the narrowed model
// (task/node1-turns-every-round): a tilt-index arrival changes ONLY this node's own
// r-index, never its angular ray, and lands the resulting distance to the partner within
// half an r-step of the model's D — for both a RADIAL partner (same iPhi) and a
// NON-RADIAL one (different iPhi, the case the naive "partnerIR - steps" shortcut gets
// wrong). Pure function, no goroutine (docs/testing-shape.md).
func TestSolveTiltIndexROffsetDistanceOnly(t *testing.T) {
	sceneCenter := vec3{X: 10, Y: 5, Z: -3}
	selfOffset := quantizedOffset{iTheta: 10, iPhi: 0, iR: 5}
	_, _, rStep := selfOffset.effectiveSteps()

	cases := []struct {
		name          string
		partnerOffset quantizedOffset
	}{
		{"radial partner (same iPhi as self)", quantizedOffset{iTheta: 10, iPhi: 0, iR: 20}},
		{"non-radial partner (different iPhi)", quantizedOffset{iTheta: 10, iPhi: 3, iR: 20}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			partnerCenter := sceneCenter.Add(polar2cart(offsetScenePolar(c.partnerOffset)))
			const thetaIdx int32 = 4
			const selfKind, partnerKind = "Node2", "Node1"

			newOffset, ok := solveTiltIndexROffset(selfOffset, sceneCenter, partnerCenter, thetaIdx, selfKind, partnerKind)
			if !ok {
				t.Fatalf("solveTiltIndexROffset returned ok=false, want a real root")
			}

			// Angles untouched.
			if newOffset.iTheta != selfOffset.iTheta || newOffset.iPhi != selfOffset.iPhi {
				t.Fatalf("angular indices changed: got (iTheta=%d,iPhi=%d), want (iTheta=%d,iPhi=%d)",
					newOffset.iTheta, newOffset.iPhi, selfOffset.iTheta, selfOffset.iPhi)
			}

			idx := thetaIdx
			if idx < 0 {
				idx = -idx
			}
			steps := int(idx) + nodeTorusSteps(selfKind) + nodeTorusSteps(partnerKind)
			wantD := float64(steps) * wire.BeadStepR

			newPos := sceneCenter.Add(polar2cart(offsetScenePolar(newOffset)))
			gotD := newPos.Sub(partnerCenter).Length()
			if math.Abs(gotD-wantD) > rStep/2 {
				t.Fatalf("distance to partner = %.4f, want within %.4f of D=%.4f (iR=%d)", gotD, rStep/2, wantD, newOffset.iR)
			}
		})
	}
}

// TestSolveTiltIndexROffsetKeepsCurrentSide pins the "never jump to the far side of the
// scene centre" guard: of the quadratic's two roots, the one chosen must be the one
// nearer this node's CURRENT radius, not merely the first algebraic root.
func TestSolveTiltIndexROffsetKeepsCurrentSide(t *testing.T) {
	sceneCenter := vec3{}
	// A partner placed so the two roots of the quadratic are on clearly different sides
	// (near vs. far) of the node's current radius.
	selfOffset := quantizedOffset{iTheta: 0, iPhi: 0, iR: 30}
	partnerOffset := quantizedOffset{iTheta: 0, iPhi: 0, iR: 2}
	partnerCenter := sceneCenter.Add(polar2cart(offsetScenePolar(partnerOffset)))

	const thetaIdx int32 = 1
	const selfKind, partnerKind = "Node2", "Node1"

	newOffset, ok := solveTiltIndexROffset(selfOffset, sceneCenter, partnerCenter, thetaIdx, selfKind, partnerKind)
	if !ok {
		t.Fatalf("solveTiltIndexROffset returned ok=false, want a real root")
	}
	// Along this radial ray the two roots are symmetric about the partner's own
	// radius (partnerOffset.iR*rStep ± D/rStep): the near one stays on the same side
	// (>= partner's radius, since self started far beyond it) as the current radius.
	if newOffset.iR < partnerOffset.iR {
		t.Fatalf("chosen root iR=%d crossed to the far side of the partner (started at iR=%d, partner iR=%d)",
			newOffset.iR, selfOffset.iR, partnerOffset.iR)
	}
}
