package Wiring

import (
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
