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
	// Tolerance, not an exact-zero Mod check: every caller builds centerGap as
	// `cells * wire.DefaultLocalStepR` so the intent is always an exact cell count, but
	// wire.DefaultLocalStepR (2.24) is not exactly representable in float64, so the
	// multiplication itself accrues ~1e-14 of round-off before this function ever sees it.
	// A strict Mod-must-be-zero check treated that round-off as a caller bug; it isn't one.
	quantIRf := centerGap / wire.DefaultLocalStepR
	if math.Abs(quantIRf-math.Round(quantIRf)) > 1e-6 {
		panic("singleNeighborHolder: centerGap must be an exact multiple of the default stepR")
	}
	quantIR := int(math.Round(quantIRf))
	lh := &wire.LayoutHolder{}
	lh.SetLocalPolar(to, 0, 0, quantIR, 0, 0, 0)
	return func() *wire.LayoutHolder { return lh }
}

// The invariant the original tests missed, and the bug that shipped: a bead must never sit
// inside either node's sphere — restated in TANGENCY terms under the bead lattice
// (docs/bead-lattice.md "Placement"): bead 0's torus is tangent OUTSIDE the source node's
// torus, and the last bead's torus is tangent OUTSIDE the target's, never overlapping either.
func TestChainBeadsStayOutsideBothNodes(t *testing.T) {
	// Expressed as a cell count * wire.DefaultLocalStepR, not a bare literal, so this
	// stays an exact multiple of the local-polar grid constant whatever that constant is
	// (it changed from 2.0 to 2.24 when the bead radius became the authored primitive —
	// docs/bead-lattice.md "The lattice is derived, not the bead").
	const gap = 200 * wire.DefaultLocalStepR
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
	srcClear := nodeTorusOuterR("Input")
	dstClear := gap - nodeTorusOuterR("Time")
	for i := range ox {
		d := math.Sqrt(float64(ox[i]*ox[i] + oy[i]*oy[i] + oz[i]*oz[i]))
		// Tangent-outside at each end: a bead's OWN torus clears the surface too, so no bead
		// is even half-buried.
		if d < srcClear-1e-4 {
			t.Errorf("bead %d at %.3f from center is inside the SOURCE node (needs >= %.3f)", i, d, srcClear)
		}
		if d > dstClear+1e-4 {
			t.Errorf("bead %d at %.3f from center is inside the TARGET node (needs <= %.3f)", i, d, dstClear)
		}
	}
}

// Beads TOUCH: adjacent centers are exactly one bead-lattice STEP apart
// (wire.BeadStepR), so a chain is a solid line with no gaps.
func TestChainBeadsTouch(t *testing.T) {
	m := &nodeMover{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
		outTargets: []string{"b"}, cascadeKinds: map[string]string{"b": "Input"},
		// 150 cells * wire.DefaultLocalStepR, not the bare literal 300 — see the comment
		// on TestChainBeadsStayOutsideBothNodes's gap.
		layoutHolderFn: singleNeighborHolder("b", 150*wire.DefaultLocalStepR),
	}
	ox, oy, oz, _, _ := m.chainBeads()
	if len(ox) < 3 {
		t.Fatalf("want several beads to compare spacing, got %d", len(ox))
	}
	want := wire.BeadStepR
	for i := 1; i < len(ox); i++ {
		dx, dy, dz := ox[i]-ox[i-1], oy[i]-oy[i-1], oz[i]-oz[i-1]
		got := math.Sqrt(float64(dx*dx + dy*dy + dz*dz))
		if math.Abs(got-want) > 1e-3 {
			t.Errorf("beads %d..%d are %.4f apart, want one bead step %.4f (touching, no gap)", i-1, i, got, want)
		}
	}
}

// Two nodes whose TORI are closer than one bead step apart still contribute at least
// ONE bead — edgeStepCount clamps to a minimum of 1 (docs/bead-lattice.md "The count") so an
// edge never has zero beads, even when the gap collapses.
func TestChainBeadsAlwaysAtLeastOneBead(t *testing.T) {
	m := &nodeMover{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
		outTargets: []string{"b"}, cascadeKinds: map[string]string{"b": "Input"},
		// 16 cells * wire.DefaultLocalStepR, not the bare literal 32 — see the comment on
		// TestChainBeadsStayOutsideBothNodes's gap.
		layoutHolderFn: singleNeighborHolder("b", 16*wire.DefaultLocalStepR),
	}
	if ox, _, _, _, _ := m.chainBeads(); len(ox) != 1 {
		t.Errorf("count = %d, want 1 — edgeStepCount clamps a collapsed gap to the minimum, never 0", len(ox))
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

// Count is proportional to the GAP (torus-to-torus, edgeStepCount's `gap`) — not to the
// center distance, which is what the pre-bead-lattice version asserted and what let the
// buried-bead bug through. Double the span, double the beads (±1 for the round), which is
// what makes a constant per-bead dwell a constant visible speed.
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
	// base and span must each be an exact multiple of wire.DefaultLocalStepR
	// (singleNeighborHolder's requirement — a sum of two such multiples is one too, so
	// base+span and base+2*span stay exact) and base must be comfortably clear of
	// 2*nodeTorusOuterR("Input") so the smaller span still produces beads.
	const base = 20 * wire.DefaultLocalStepR
	span := 80 * wire.DefaultLocalStepR
	n1 := count(base + span)
	n2 := count(base + 2*span)
	if n2 < 2*n1-1 || n2 > 2*n1+1 {
		t.Errorf("count(span %.0f) = %d, count(span %.0f) = %d; want roughly double", span, n1, 2*span, n2)
	}
}

// THE REGRESSION TEST for the reported bug: every bead index must be reachable as traversal
// progress t sweeps the whole edge. The bug was two coordinate systems — lighting quantised
// distance against the wire's PORT-TO-PORT arc while the chain was laid out to a longer
// center-distance-minus-radii span — so the tail of the chain could never be lit, however far
// t climbed. Under the bead lattice both layout and lighting read the SAME integer step
// count, so this now sweeps t directly against that integer.
func TestChainBeadsEveryIndexIsReachable(t *testing.T) {
	const steps = 28 // an arbitrary, but multi-bead, step count
	seen := make(map[int]bool, steps)
	const sweeps = 100000
	for i := 0; i <= sweeps; i++ {
		tt := float64(i) / float64(sweeps)
		if tt >= 1 {
			tt = math.Nextafter(1, 0) // sweep right up to, but not touching, t=1
		}
		idx, ok := litBeadIndex(tt, steps)
		if !ok {
			continue
		}
		seen[idx] = true
	}
	for i := 0; i < steps; i++ {
		if !seen[i] {
			t.Errorf("bead index %d of %d was never reached while sweeping t in [0,1) — unreachable tail bead", i, steps)
		}
	}
	if len(seen) != steps {
		t.Errorf("saw %d distinct indices, want exactly %d (0..steps-1)", len(seen), steps)
	}
}

// The invariant two versions of litBeadIndex violated: two beads placed in ONE emission travel at
// the same world speed, so after the same ELAPSED time they must light the same bead index —
// whatever their edges' STEP COUNTS. Node 1's two edges differ in length (hence step count) —
// give them a different ratio to each other than a naive proportional guess, so a version
// that gets the per-edge scaling wrong fails.
//
// Driving this by elapsed time rather than by a chosen distance is the point: it is what the
// screen shows, and it is what caught t*centerDistance / t*chordLength, where a per-edge
// ratio reintroduced an offset that t*steps does not have (dwell is UNIFORM per step, so
// elapsed/dwell is the same integer progress for every edge regardless of its own step count).
func TestLitBeadIndexSameElapsedLightsSameBead(t *testing.T) {
	const longSteps, shortSteps = 32, 28

	for elapsed := 0.0; elapsed < 120; elapsed += 0.25 {
		coveredSteps := elapsed / wire.DwellTicksPerBead // elapsed is in the SAME ticks unit as dwell
		gotLong, okLong := litBeadIndex(coveredSteps/longSteps, longSteps)
		gotShort, okShort := litBeadIndex(coveredSteps/shortSteps, shortSteps)
		if !okLong || !okShort {
			continue
		}
		if gotLong != gotShort {
			t.Fatalf("elapsed %.2f (covered %.2f steps): long edge lit bead %d, short edge lit bead %d — equal elapsed must light the same index",
				elapsed, coveredSteps, gotLong, gotShort)
		}
	}
}

// One bead index per dwell of travel — the constant dwell the design rests on. If this
// drifts, the lit bead is no longer moving at the uniform pulse speed.
func TestLitBeadIndexAdvancesOncePerStep(t *testing.T) {
	const steps = 25
	for i := 0; i < steps; i++ {
		got, ok := litBeadIndex(float64(i)/steps, steps)
		if !ok {
			t.Fatalf("bead %d: t=%.4f reported off-chain", i, float64(i)/steps)
		}
		if got != i {
			t.Errorf("t=%.4f (bead %d's own position) lit index %d, want %d", float64(i)/steps, i, got, i)
		}
	}
}

// t outside [0,1) — before departure, or having arrived — lights nothing.
func TestLitBeadIndexOffChainLightsNothing(t *testing.T) {
	const steps = 25
	if _, ok := litBeadIndex(-0.01, steps); ok {
		t.Error("t<0 (not yet departed) lit a bead; want nothing lit")
	}
	if _, ok := litBeadIndex(1, steps); ok {
		t.Error("t=1 (arrived at the target) lit a bead; want nothing lit")
	}
}

// edgeStepCount pins the formula (docs/bead-lattice.md "The count") directly, independent of
// chainBeads' node-mover plumbing.
func TestEdgeStepCount(t *testing.T) {
	lp := wire.LocalPolar{QuantIR: 200} // separation = 200 * wire.DefaultLocalStepR
	got := edgeStepCount(lp, "Input", "Time")
	gap := 200*wire.DefaultLocalStepR - nodeTorusOuterR("Input") - nodeTorusOuterR("Time")
	want := int(math.Round(gap / wire.BeadStepR))
	if got != want {
		t.Fatalf("edgeStepCount = %d, want %d", got, want)
	}
	if want < 1 {
		t.Fatal("test fixture collapsed to <1 step; pick a larger separation")
	}
}

// A collapsed or negative gap clamps to a minimum of 1 bead — an edge is never zero-length.
func TestEdgeStepCountClampsToMinimumOne(t *testing.T) {
	lp := wire.LocalPolar{QuantIR: 1} // separation = wire.DefaultLocalStepR, far inside both tori
	if got := edgeStepCount(lp, "Input", "Time"); got != 1 {
		t.Fatalf("edgeStepCount(collapsed) = %d, want 1", got)
	}
}

// THE REGRESSION GUARD for this commit: exact double tangency, no float tolerance wider
// than round-trip noise. Before this commit, edgeStepCount measured
// `round((QuantIR*stepR - nodeTorusOuterR(src) - nodeTorusOuterR(dst)) / BeadStepR)` against
// an nodeTorusOuterR that was an arbitrary float NOT on the bead lattice
// (nodeRadius(kind)*(1+ShadingParamNodeRingTubeRatio)), so the division was essentially never
// exact and the rounding silently absorbed up to half a bead step at the TARGET end — bead 0
// was always exactly tangent to the source (offset by construction), but the last bead's far
// edge only APPROXIMATELY met the target's torus. This test pins the far edge to the target's
// torus to float-round-off tolerance (1e-6, not the 1e-4 "stays outside" tolerance the older
// tests use, because this is asserting equality, not clearance), across several QuantIR values
// (already snapped to the bead lattice, matching what LayoutHolder.SetLocalPolar actually
// stores — see singleNeighborHolder) and several node-kind pairs whose bareNodeRadius values
// don't share an obvious common factor. Reintroducing a float `gap/BeadStepR` division against
// an unsnapped nodeTorusOuterR reopens exactly the discrepancy this test would catch: verified
// by hand during development by temporarily restoring the pre-fix
// `nodeRadius(kind)*(1+ShadingParamNodeRingTubeRatio)` definition of nodeTorusOuterR and
// `int(math.Round(gap/wire.BeadStepR))` in edgeStepCount, which failed this test with a
// last-bead far-edge error up to half of wire.BeadStepR (4.0 world units) — see this commit's
// message for the exact numbers.
// tangencyEps is the equality tolerance for this test's near/far edge assertions —
// 1e-6 (round-trip float64 noise) before the primitive/derived flip, widened to 1e-3
// after it. That is not a weakening of the invariant: chainBeads' ox/oy/oz are float32
// (the streamed buffer type), and this test's positions reach the O(1000) world-unit
// range, where float32's ~7 significant digits give ~1e-4 of unavoidable rounding per
// coordinate regardless of geometry. Under the OLD LocalStepR=2.0 those positions
// happened to be exactly float32-representable and this noise never showed; under the
// NEW bead-authored LocalStepR=2.24 (docs/bead-lattice.md "The lattice is derived, not
// the bead") they generally are not, so the noise is now visible. 1e-3 is still four
// orders of magnitude tighter than "half a bead step" (wire.BeadStepR/2 = 4.48), the
// error this test exists to catch (a reintroduced math.Round division).
const tangencyEps = 1e-3

func TestChainBeadsExactDoubleTangency(t *testing.T) {
	kindPairs := [][2]string{
		{"Input", "Time"},
		{"Time", "Input"},
		{"Input", "Input"},
		{"Time", "Time"},
	}
	// Every centerGap here must be an exact multiple of wire.DefaultLocalStepR
	// (singleNeighborHolder's own requirement) and large enough to clear both tori with
	// room for at least one bead, across the largest kind pair tested. Expressed as cell
	// counts * wire.DefaultLocalStepR, not bare literals, so this stays exact whatever
	// that constant is (see TestChainBeadsStayOutsideBothNodes's gap comment).
	cellCounts := []float64{100, 120, 180, 260, 500}
	centerGaps := make([]float64, len(cellCounts))
	for i, c := range cellCounts {
		centerGaps[i] = c * wire.DefaultLocalStepR
	}

	for _, kp := range kindPairs {
		srcKind, dstKind := kp[0], kp[1]
		for _, gap := range centerGaps {
			m := &nodeMover{
				id:             "a",
				geom:           nodeGeom{nodeIdentity: nodeIdentity{Kind: srcKind}},
				outTargets:     []string{"b"},
				cascadeKinds:   map[string]string{"b": dstKind},
				layoutHolderFn: singleNeighborHolder("b", gap),
			}
			ox, oy, oz, _, _ := m.chainBeads()
			if len(ox) == 0 {
				t.Fatalf("%s->%s gap %.0f: no beads emitted", srcKind, dstKind, gap)
			}

			srcTorus := nodeTorusOuterR(srcKind)
			dstTorus := nodeTorusOuterR(dstKind)

			// Bead 0's NEAR edge: its offset from center minus its own torus radius must
			// equal the source's torus radius EXACTLY — this held even before this
			// commit (it's placement by direct addition, not a derived count), pinned
			// here so a future change can't break it while "fixing" the far end.
			d0 := math.Sqrt(float64(ox[0])*float64(ox[0]) + float64(oy[0])*float64(oy[0]) + float64(oz[0])*float64(oz[0]))
			nearEdge0 := d0 - wire.BeadTorusOuterR
			if math.Abs(nearEdge0-srcTorus) > tangencyEps {
				t.Errorf("%s->%s gap %.0f: bead 0 near edge %.9f, want exactly srcTorus %.9f",
					srcKind, dstKind, gap, nearEdge0, srcTorus)
			}

			// Last bead's FAR edge: its offset from center plus its own torus radius must
			// equal the target's torus's NEAR edge (separation - dstTorus) exactly — the
			// invariant that used to be a rounding coincidence.
			last := len(ox) - 1
			dLast := math.Sqrt(float64(ox[last])*float64(ox[last]) + float64(oy[last])*float64(oy[last]) + float64(oz[last])*float64(oz[last]))
			farEdgeLast := dLast + wire.BeadTorusOuterR
			wantFarEdge := gap - dstTorus
			if math.Abs(farEdgeLast-wantFarEdge) > tangencyEps {
				t.Errorf("%s->%s gap %.0f: last bead (%d) far edge %.9f, want exactly target's near torus edge %.9f (off by %.9f, up to half a bead step %.3f would indicate math.Round is back)",
					srcKind, dstKind, gap, last, farEdgeLast, wantFarEdge, farEdgeLast-wantFarEdge, wire.BeadStepR/2)
			}
		}
	}
}
