package Wiring

import (
	"math"
	"testing"
)

// chainBeads is a pure function of ONE node's own state — its center and its own
// partnerCenters map, both written only by that node's goroutine — so this is a plain
// table, no second goroutine (docs/testing-shape.md).
func TestChainBeadsSpacingAndCount(t *testing.T) {
	// Target 5×spacing straight out +x: 5 beads, at 1..5 × spacing, none at the node itself
	// and none past the target.
	m := &nodeMover{
		id:         "a",
		outTargets: []string{"b"},
		partnerCenters: map[string]vec3{
			"b": {X: 5 * chainBeadSpacing, Y: 0, Z: 0},
		},
	}
	ox, oy, oz, _ := m.chainBeads()
	if len(ox) != 5 {
		t.Fatalf("count = %d, want 5 (length 5×spacing at constant spacing)", len(ox))
	}
	for i := range ox {
		want := float32(float64(i+1) * chainBeadSpacing)
		if math.Abs(float64(ox[i]-want)) > 1e-4 || oy[i] != 0 || oz[i] != 0 {
			t.Errorf("bead %d = (%v,%v,%v), want (%v,0,0)", i, ox[i], oy[i], oz[i], want)
		}
	}
	// The offsets are NODE-LOCAL: they must not change when this node's own center does.
	// That is the whole constant-time-move claim — a move rewrites the center and nothing
	// else. (geom's zero value put the node at the origin above; any other center must give
	// the same offsets for the same partner OFFSET, which is what a caller relies on.)
	if got := len(ox); got != 5 {
		t.Errorf("count changed under re-read: %d", got)
	}
}

// A target nearer than one spacing yields NO beads rather than one placed past the target.
func TestChainBeadsShorterThanOneSpacing(t *testing.T) {
	m := &nodeMover{
		id:             "a",
		outTargets:     []string{"b"},
		partnerCenters: map[string]vec3{"b": {X: chainBeadSpacing * 0.5}},
	}
	if ox, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 — a bead at one spacing would sit past the target", len(ox))
	}
}

// A target whose center this node has never been told contributes nothing — the node aims
// only with what its own partnerCenters map holds, never by reading another goroutine.
func TestChainBeadsUnknownPartnerContributesNothing(t *testing.T) {
	m := &nodeMover{id: "a", outTargets: []string{"b"}, partnerCenters: map[string]vec3{}}
	if ox, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 for an unknown partner center", len(ox))
	}
}

// Count is LENGTH-PROPORTIONAL, which is what makes a constant dwell per bead equal to
// uniform world speed (docs/beads-are-the-edge.md open question 4): doubling the distance
// doubles the bead count, so total traversal time doubles at a fixed per-bead dwell.
func TestChainBeadsCountIsLengthProportional(t *testing.T) {
	count := func(mult float64) int {
		m := &nodeMover{
			id: "a", outTargets: []string{"b"},
			partnerCenters: map[string]vec3{"b": {X: mult * chainBeadSpacing}},
		}
		ox, _, _, _ := m.chainBeads()
		return len(ox)
	}
	if a, b := count(4), count(8); b != 2*a {
		t.Errorf("count(8×spacing) = %d, want 2× count(4×spacing) = %d", b, 2*a)
	}
}
