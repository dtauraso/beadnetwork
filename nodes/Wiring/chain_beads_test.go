package Wiring

import (
	"math"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// chainBeads is a pure function of ONE node's own state — its own kind, its own
// cascadeKinds, and its own stored LocalPolar list (m.layoutHolderFn) — all written only by
// that node's goroutine — so these are plain tables, no second goroutine
// (docs/testing-shape.md).
//
// None of these tests wire up outWireOuts, so chainBeads always takes the FALLBACK arc
// (neighborDist - selfRadius - targetRadius, the local surface-to-surface estimate — see
// chainBeads' arc-source doc comment). That fallback is exactly what makes these plain
// single-node tables: no *wire.Out, no clock-driven geometry, nothing beyond this node's own
// fields.
//
// Every case sets BOTH kinds explicitly. An unset kind is NOT neutral: kindWidthHeight falls
// back to (110, 60), i.e. radius 15. The original tests left kinds unset and asserted on
// distance from the node CENTER, which is exactly why they passed while beads rendered inside
// the nodes.
//
// A neighbour's position is supplied the same way the production code reads it: as a stored
// LocalPolar (QuantITheta/QuantIPhi/QuantIR, default step constants) on a bare *wire.LayoutHolder,
// handed back through layoutHolderFn — never as a cartesian partnerCenters offset. Distance is
// therefore an EXACT index multiple of the default stepR (2.0 world units,
// nodes/wire/layout_holder.go), so every "centerGap" below is chosen to be even.

// singleNeighborHolder builds a bare LayoutHolder holding exactly one LocalPolar entry — this
// node's stored bearing/distance to `to` — at QuantITheta=QuantIPhi=0 (an arbitrary but fixed
// direction; magnitude-only assertions below don't care which one) and QuantIR chosen so that
// float64(QuantIR)*wire.DefaultLocalStepR == centerGap.
func singleNeighborHolder(to string, centerGap float64) func() *wire.LayoutHolder {
	if math.Mod(centerGap, wire.DefaultLocalStepR) != 0 {
		panic("singleNeighborHolder: centerGap must be an exact multiple of the default stepR")
	}
	quantIR := int(math.Round(centerGap / wire.DefaultLocalStepR))
	lh := &wire.LayoutHolder{}
	lh.SetLocalPolar(to, 0, 0, quantIR, 0, 0, 0)
	return func() *wire.LayoutHolder { return lh }
}

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
		layoutHolderFn: singleNeighborHolder("b", gap),
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
		// is even half-buried. (gap=400 with radii 15/9 gives arc=376, an exact multiple of
		// chainBeadSpacing=8, so the last bead lands exactly tangent rather than merely
		// "somewhere clear" — floor(arc/spacing) guarantees count*spacing <= arc always, see
		// beadCount's doc comment.)
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
		layoutHolderFn: singleNeighborHolder("b", 300),
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

// Two nodes whose SURFACES are closer than one bead diameter contribute NO beads, rather than
// a bead that can't fully fit in the gap. Centers 32 apart with two radius-15 nodes leaves an
// arc of 2, well under chainBeadSpacing (8).
func TestChainBeadsNoneWhenSurfacesTooClose(t *testing.T) {
	m := &nodeMover{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
		outTargets: []string{"b"}, cascadeKinds: map[string]string{"b": "Input"},
		layoutHolderFn: singleNeighborHolder("b", 32),
	}
	if ox, _, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 — no bead fits between the two surfaces", len(ox))
	}
}

// A target this node has no stored local polar to contributes nothing — the node aims only
// with what its own LayoutHolder holds, never by reading another goroutine.
func TestChainBeadsUnknownPartnerContributesNothing(t *testing.T) {
	lh := &wire.LayoutHolder{}
	m := &nodeMover{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}}, outTargets: []string{"b"},
		layoutHolderFn: func() *wire.LayoutHolder { return lh },
	}
	if ox, _, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 for an unknown partner", len(ox))
	}
}

// No LayoutHolder at all (layoutHolderFn nil, or returning nil) contributes nothing rather
// than panicking — the same "no cross-goroutine read, no made-up direction" contract as an
// unknown partner.
func TestChainBeadsNoLayoutHolderContributesNothing(t *testing.T) {
	m := &nodeMover{id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}}, outTargets: []string{"b"}}
	if ox, _, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 with no layoutHolderFn", len(ox))
	}
}

// Count is proportional to the ARC (the surface-to-surface span in the fallback case) — not to
// the center distance, which is what the previous version asserted and what let the
// buried-bead bug through. Double the span, double the beads (±1 for the floor), which is what
// makes a constant per-bead dwell a constant visible speed.
func TestChainBeadsCountIsSpanProportional(t *testing.T) {
	count := func(centerGap float64) int {
		m := &nodeMover{
			id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
			outTargets: []string{"b"}, cascadeKinds: map[string]string{"b": "Input"},
			layoutHolderFn: singleNeighborHolder("b", centerGap),
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

// THE REGRESSION TEST for the reported bug: every bead index must be reachable as traversal
// progress t sweeps the whole edge. The bug was two coordinate systems — lighting quantised
// distance against the wire's PORT-TO-PORT arc while the chain was laid out to a longer
// center-distance-minus-radii span — so the tail of the chain (here: the last 1-2 beads) could
// never be lit, however far t climbed. Sweeping d from 0 to arc and demanding the union of hit
// indices is exactly {0, ..., count-1} makes that failure mode fail loudly instead of shipping
// silently: an unreachable index leaves a gap in the observed set.
func TestChainBeadsEveryIndexIsReachable(t *testing.T) {
	const arc = 224.0 // matches the measured node-1 1->3 arc from the bug report
	count := beadCount(arc)
	if count == 0 {
		t.Fatal("test arc produced zero beads; pick a larger arc")
	}
	seen := make(map[int]bool, count)
	const steps = 100000
	for i := 0; i <= steps; i++ {
		d := arc * float64(i) / float64(steps)
		t := d / arc
		if t >= 1 {
			t = math.Nextafter(1, 0) // sweep right up to, but not touching, t=1
		}
		idx, ok := litBeadIndex(t, arc)
		if !ok {
			continue
		}
		seen[idx] = true
	}
	for i := 0; i < count; i++ {
		if !seen[i] {
			t.Errorf("bead index %d of %d was never reached while sweeping t in [0,1) — unreachable tail bead", i, count)
		}
	}
	if len(seen) != count {
		t.Errorf("saw %d distinct indices, want exactly %d (0..count-1)", len(seen), count)
	}
}

// The invariant two versions of litBeadIndex violated: two beads placed in ONE emission travel at
// the same world speed, so after the same ELAPSED time they must light the same bead index —
// whatever their edges' lengths. Node 1's edges differ 1.9% (259.208 vs 254.334 measured) and
// both chains hold 28 beads, so a rate error shows up as one bead permanently ahead.
//
// Driving this by elapsed time rather than by a chosen distance is the point: it is what the
// screen shows, and it is what caught t*centerDistance, where the per-edge (center/arc) ratio
// reintroduced the offset that t*beadCount had.
func TestLitBeadIndexSameElapsedLightsSameBead(t *testing.T) {
	// Arcs are port-to-port and are NOT the center separations; give them a different ratio to
	// each other than the centers have, so a version that multiplies by the wrong length fails.
	longArc, shortArc := 259.208, 254.334

	const pulseSpeed = 1.7 // any positive speed: it is shared by every wire
	for elapsed := 0.0; elapsed < 120; elapsed += 0.25 {
		covered := elapsed * pulseSpeed
		gotLong, okLong := litBeadIndex(covered/longArc, longArc)
		gotShort, okShort := litBeadIndex(covered/shortArc, shortArc)
		if !okLong || !okShort {
			continue
		}
		if gotLong != gotShort {
			t.Fatalf("elapsed %.2f (covered %.2f): long edge lit bead %d, short edge lit bead %d — equal elapsed must light the same index",
				elapsed, covered, gotLong, gotShort)
		}
	}
}

// One bead index per chainBeadSpacing of travel — the constant dwell the design rests on. If
// this drifts, the lit bead is no longer moving at the uniform pulse speed.
func TestLitBeadIndexAdvancesOncePerSpacing(t *testing.T) {
	const arc = 400.0
	count := beadCount(arc)
	for i := 0; i < count; i++ {
		covered := float64(i) * chainBeadSpacing
		got, ok := litBeadIndex(covered/arc, arc)
		if !ok {
			t.Fatalf("bead %d: covered %.2f reported off-chain", i, covered)
		}
		if got != i {
			t.Errorf("covered %.2f (bead %d's own position) lit index %d, want %d", covered, i, got, i)
		}
	}
}

// t outside [0,1) — before departure, or having arrived — lights nothing.
func TestLitBeadIndexOffChainLightsNothing(t *testing.T) {
	const arc = 400.0
	if _, ok := litBeadIndex(-0.01, arc); ok {
		t.Error("t<0 (not yet departed) lit a bead; want nothing lit")
	}
	if _, ok := litBeadIndex(1, arc); ok {
		t.Error("t=1 (arrived at the target) lit a bead; want nothing lit")
	}
}
