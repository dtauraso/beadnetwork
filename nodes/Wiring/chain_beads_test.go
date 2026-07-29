package Wiring

import (
	"math"
	"testing"
)

// chainBeads is a pure function of ONE node's own state — its center, its own kind, its own
// cascadeKinds and its own partnerCenters map, all written only by that node's goroutine — so
// these are plain tables, no second goroutine (docs/testing-shape.md).
//
// Every case sets BOTH kinds explicitly. An unset kind is NOT neutral: kindWidthHeight falls
// back to (110, 60), i.e. radius 15. The original tests left kinds unset and asserted on
// distance from the node CENTER, which is exactly why they passed while beads rendered inside
// the nodes.

// The invariant the original tests missed, and the bug that shipped: a bead must never sit
// inside either node's sphere. Beads were placed by distance from the source center starting at
// one spacing (8) — inside a radius-15 node — and the count ran to the target's center, so the
// far end was buried too.
func TestChainBeadsStayOutsideBothNodes(t *testing.T) {
	const gap = 400.0
	m := &nodeMover{
		id:             "a",
		geom:           nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}}, // radius 15
		outTargets:     []string{"b"},
		cascadeKinds:   map[string]string{"b": "Time"}, // radius 9
		partnerCenters: map[string]vec3{"b": {X: gap}},
	}
	ox, oy, oz, _, _ := m.chainBeads()
	if len(ox) == 0 {
		t.Fatal("no beads emitted for a 400-unit gap")
	}
	srcClear := nodeRadius("Input") + ShadingParamBeadRadius
	dstClear := gap - nodeRadius("Time") - ShadingParamBeadRadius
	for i := range ox {
		d := math.Sqrt(float64(ox[i]*ox[i] + oy[i]*oy[i] + oz[i]*oz[i]))
		// Tangent-outside at each end: a bead's OWN radius clears the surface too, so no bead
		// is even half-buried.
		if d < srcClear-1e-4 {
			t.Errorf("bead %d at %.3f from center is inside the SOURCE node (needs >= %.3f)", i, d, srcClear)
		}
		if d > dstClear+1e-4 {
			t.Errorf("bead %d at %.3f from center is inside the TARGET node (needs <= %.3f)", i, d, dstClear)
		}
	}
}

// Beads TOUCH: adjacent centers are exactly one diameter apart, so a chain is a solid line with
// no gaps. Spacing is derived from the shared bead radius, so this pins that derivation too.
func TestChainBeadsTouch(t *testing.T) {
	m := &nodeMover{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
		outTargets: []string{"b"}, cascadeKinds: map[string]string{"b": "Input"},
		partnerCenters: map[string]vec3{"b": {X: 300}},
	}
	ox, oy, oz, _, _ := m.chainBeads()
	if len(ox) < 3 {
		t.Fatalf("want several beads to compare spacing, got %d", len(ox))
	}
	want := 2 * ShadingParamBeadRadius
	for i := 1; i < len(ox); i++ {
		dx, dy, dz := ox[i]-ox[i-1], oy[i]-oy[i-1], oz[i]-oz[i-1]
		got := math.Sqrt(float64(dx*dx + dy*dy + dz*dz))
		if math.Abs(got-want) > 1e-3 {
			t.Errorf("beads %d..%d are %.4f apart, want one diameter %.4f (touching, no gap)", i-1, i, got, want)
		}
	}
}

// Two nodes whose SURFACES are closer than one bead contribute NO beads, rather than beads
// buried in one node or the other. Centers 32 apart with two radius-15 nodes leaves 2 units.
func TestChainBeadsNoneWhenSurfacesTooClose(t *testing.T) {
	m := &nodeMover{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
		outTargets: []string{"b"}, cascadeKinds: map[string]string{"b": "Input"},
		partnerCenters: map[string]vec3{"b": {X: 32}},
	}
	if ox, _, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 — no bead fits between the two surfaces", len(ox))
	}
}

// A target whose center this node has never been told contributes nothing — the node aims only
// with what its own partnerCenters map holds, never by reading another goroutine.
func TestChainBeadsUnknownPartnerContributesNothing(t *testing.T) {
	m := &nodeMover{id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}}, outTargets: []string{"b"}, partnerCenters: map[string]vec3{}}
	if ox, _, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 for an unknown partner center", len(ox))
	}
}

// Count is proportional to the SPAN between the surfaces — not to the center distance, which is
// what the previous version asserted and what let the buried-bead bug through. Double the span,
// double the beads (±1 for the floor), which is what makes a constant per-bead dwell a constant
// visible speed.
func TestChainBeadsCountIsSpanProportional(t *testing.T) {
	count := func(centerGap float64) int {
		m := &nodeMover{
			id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
			outTargets: []string{"b"}, cascadeKinds: map[string]string{"b": "Input"},
			partnerCenters: map[string]vec3{"b": {X: centerGap}},
		}
		ox, _, _, _, _ := m.chainBeads()
		return len(ox)
	}
	clear := 2 * (nodeRadius("Input") + ShadingParamBeadRadius) // swallowed by the two nodes
	span := 160.0
	n1 := count(clear + span)
	n2 := count(clear + 2*span)
	if n2 < 2*n1-1 || n2 > 2*n1+1 {
		t.Errorf("count(span %.0f) = %d, count(span %.0f) = %d; want roughly double", span, n1, 2*span, n2)
	}
}
