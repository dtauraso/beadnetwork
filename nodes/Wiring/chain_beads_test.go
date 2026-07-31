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
	ox, oy, oz, _, _, _ := m.chainBeads()
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
	ox, oy, oz, _, _, _ := m.chainBeads()
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
		// 3 cells * wire.DefaultLocalStepR, not the bare literal 6 — see the comment on
		// TestChainBeadsStayOutsideBothNodes's gap. QuantIR IS the bead-step count now
		// (edgeStepCount no longer divides by a cell-per-step constant — bead_lattice.go's
		// BeadStepCells doc comment), so 3 steps minus both "Input" nodes' own
		// nodeTorusSteps (2 each) collapses well below the minimum.
		layoutHolderFn: singleNeighborHolder("b", 3*wire.DefaultLocalStepR),
	}
	if ox, _, _, _, _, _ := m.chainBeads(); len(ox) != 1 {
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
	if ox, _, _, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 for an unknown partner", len(ox))
	}
}

// No LayoutHolder at all (layoutHolderFn nil, or returning nil) contributes nothing rather
// than panicking — the same "no cross-goroutine read, no made-up direction" contract as an
// unknown partner.
func TestChainBeadsNoLayoutHolderContributesNothing(t *testing.T) {
	m := &nodeMover{id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}}, outTargets: []string{"b"}}
	if ox, _, _, _, _, _ := m.chainBeads(); len(ox) != 0 {
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
		ox, _, _, _, _, _ := m.chainBeads()
		return len(ox)
	}
	// base and span must each be an exact multiple of wire.DefaultLocalStepR
	// (singleNeighborHolder's requirement — a sum of two such multiples is one too, so
	// base+span and base+2*span stay exact). base is chosen so it exactly cancels the
	// two nodeTorusSteps("Input") subtractions (2 each, 4 total, edgeStepCount): with
	// QuantIR itself the bead-step count (no per-step division anymore — bead_lattice.go's
	// BeadStepCells doc comment), count(base+span) = base+span-4 = span when base=4, and
	// count(base+2*span) = 2*span exactly — an exact double, not merely "roughly", because
	// there is nothing left to round.
	const base = 4 * wire.DefaultLocalStepR
	span := 80 * wire.DefaultLocalStepR
	n1 := count(base + span)
	n2 := count(base + 2*span)
	if n2 != 2*n1 {
		t.Errorf("count(span %.0f) = %d, count(span %.0f) = %d; want exactly double", span, n1, 2*span, n2)
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
// chainBeads' node-mover plumbing. QuantIR is now itself a count of bead steps (there is one
// lattice, not two — bead_lattice.go's BeadStepCells doc comment), so this is plain integer
// subtraction with nothing to round.
func TestEdgeStepCount(t *testing.T) {
	lp := wire.LocalPolar{QuantIR: 200}
	got := edgeStepCount(lp, "Input", "Time")
	want := 200 - nodeTorusSteps("Input") - nodeTorusSteps("Time")
	if got != want {
		t.Fatalf("edgeStepCount = %d, want %d", got, want)
	}
	if want < 1 {
		t.Fatal("test fixture collapsed to <1 step; pick a larger separation")
	}
}

// A collapsed or negative gap clamps to a minimum of 1 bead — an edge is never zero-length.
func TestEdgeStepCountClampsToMinimumOne(t *testing.T) {
	lp := wire.LocalPolar{QuantIR: 1} // 1 bead step of separation, far inside both tori
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
			ox, oy, oz, _, _, _ := m.chainBeads()
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

// --- "the chain spans the gap" (docs/bead-lattice.md residue, live-position fix) ---
//
// The tests above (TestChainBeadsExactDoubleTangency et al.) never trip the defect
// because singleNeighborHolder always builds a centerGap that is an EXACT multiple of
// the local-polar lattice, so count*BeadStepR already lands exactly on the real
// separation and there is no rounding residue to absorb. The reported bug only shows up
// when the two nodes' LIVE cartesian positions do NOT sit on a whole multiple of
// BeadStepR — which is the normal case once a node has been dragged — so these tests
// drive the gap through m.partnerCenters (the live cartesian map chainBeads now reads)
// instead, deliberately choosing separations that are NOT commensurate with BeadStepR.

// centerGapNotOnLattice builds a *nodeMover set up to test one edge "a"->"b" with an
// arbitrary bead COUNT (still read from the stored, quantized LocalPolar, exactly as
// edgeStepCount always has) and an ARBITRARY, independently-chosen actual surface gap
// (delivered via m.partnerCenters, the live-cartesian map applyCenter keeps current).
// selfCenter is left at the origin (m.geom.HasPos is false, so nodeWorldPos(m.geom)
// returns the zero vector) and the target is placed straight out along +X at exactly
// selfTorusR + actualGap + targetTorusR, so the surface-to-surface distance
// edgeSurfaceGap computes is exactly actualGap by construction.
func centerGapNotOnLattice(srcKind, dstKind string, quantIRForCount int, actualGap float64) *nodeMover {
	lh := &wire.LayoutHolder{}
	// The stored LocalPolar only needs to produce a bead COUNT here — its QuantIR is
	// deliberately unrelated to actualGap, which is the whole point: count and real
	// distance no longer have to agree on the same integer.
	lh.SetLocalPolar("b", 0, 0, quantIRForCount, 0, 0, 0)
	selfTorus := nodeTorusOuterR(srcKind)
	dstTorus := nodeTorusOuterR(dstKind)
	return &nodeMover{
		id:             "a",
		geom:           nodeGeom{nodeIdentity: nodeIdentity{Kind: srcKind}}, // HasPos false -> center at origin
		outTargets:     []string{"b"},
		cascadeKinds:   map[string]string{"b": dstKind},
		layoutHolderFn: func() *wire.LayoutHolder { return lh },
		partnerCenters: map[string]vec3{"b": {X: selfTorus + actualGap + dstTorus}},
	}
}

// beadOuterRFromSphereR converts a returned bead SPHERE radius back to its OUTER radius
// (the ring included) — the inverse of chainBeads' own sphereR = beadOuterR /
// (1 + wire.BeadRingTubeRatio).
func beadOuterRFromSphereR(sphereR float32) float64 {
	return float64(sphereR) * (1 + wire.BeadRingTubeRatio)
}

// chainNearFarEdges reads bead 0's near edge and the last bead's far edge (both as a
// distance from the origin — this file's fixtures always leave selfCenter at the
// origin) from chainBeads' returned offsets. Each edge's beads are now sized per edge
// (radius), so the outer radius used to convert a centre distance into a surface edge is
// read from the RETURNED radius column, not the fixed wire.BeadTorusOuterR constant.
func chainNearFarEdges(t *testing.T, ox, oy, oz, radius []float32) (near, far float64) {
	t.Helper()
	if len(ox) == 0 {
		t.Fatal("no beads emitted")
	}
	d0 := math.Sqrt(float64(ox[0])*float64(ox[0]) + float64(oy[0])*float64(oy[0]) + float64(oz[0])*float64(oz[0]))
	last := len(ox) - 1
	dLast := math.Sqrt(float64(ox[last])*float64(ox[last]) + float64(oy[last])*float64(oy[last]) + float64(oz[last])*float64(oz[last]))
	return d0 - beadOuterRFromSphereR(radius[0]), dLast + beadOuterRFromSphereR(radius[last])
}

// THE FIX's regression test: for a range of separations that are deliberately NOT whole
// multiples of BeadStepR, both ends stay exactly tangent — bead 0's near edge to the
// source's torus, the last bead's far edge to the target's torus (live-position values,
// not the quantized LocalPolar) — to float tolerance.
func TestChainBeadsTangentToLiveGapNotOnLattice(t *testing.T) {
	// Fractions of BeadStepR chosen to avoid landing back on a whole multiple by
	// coincidence (0.5 would be the exact rounding tie the bug report describes; the
	// others cover the rest of the residue range).
	fractions := []float64{0.05, 0.5, 0.9, 0.33, 0.71}
	const count = 12 // an arbitrary, but multi-bead (count>1), step count
	quantIRForCount := count + nodeTorusSteps("Input") + nodeTorusSteps("Time")

	for _, frac := range fractions {
		actualGap := (float64(count) + frac) * wire.BeadStepR
		m := centerGapNotOnLattice("Input", "Time", quantIRForCount, actualGap)
		ox, oy, oz, _, _, radius := m.chainBeads()
		if got := len(ox); got != count {
			t.Fatalf("frac %.2f: bead count = %d, want edgeStepCount's %d", frac, got, count)
		}
		near, far := chainNearFarEdges(t, ox, oy, oz, radius)
		srcTorus := nodeTorusOuterR("Input")
		wantFar := srcTorus + actualGap
		if math.Abs(near-srcTorus) > tangencyEps {
			t.Errorf("frac %.2f: bead 0 near edge %.9f, want exactly source torus %.9f", frac, near, srcTorus)
		}
		if math.Abs(far-wantFar) > tangencyEps {
			t.Errorf("frac %.2f: last bead far edge %.9f, want exactly the target's torus surface %.9f (off by %.9f)",
				frac, far, wantFar, far-wantFar)
		}
	}
}

// count still follows edgeStepCount exactly — spanning the real gap changes SPACING
// only, never how many beads there are.
func TestChainBeadsSpanFixKeepsCountFromEdgeStepCount(t *testing.T) {
	const count = 9
	quantIRForCount := count + nodeTorusSteps("Input") + nodeTorusSteps("Time")
	// A gap deliberately off the lattice by a third of a bead step.
	actualGap := (float64(count) + 0.33) * wire.BeadStepR
	m := centerGapNotOnLattice("Input", "Time", quantIRForCount, actualGap)
	ox, _, _, _, _, _ := m.chainBeads()
	want := edgeStepCount(wire.LocalPolar{QuantIR: quantIRForCount}, "Input", "Time")
	if want != count {
		t.Fatalf("test fixture: edgeStepCount = %d, want %d", want, count)
	}
	if len(ox) != want {
		t.Errorf("bead count = %d, want edgeStepCount's %d", len(ox), want)
	}
}

// Per-edge bead SIZE, not stretched spacing, absorbs the lattice-vs-live-gap residue: for
// any gap (on or off the BeadStepR lattice), consecutive bead centres are EXACTLY
// 2*beadOuterR apart — beadOuterR = gap/(2*count) — and that same beadOuterR is what the
// returned radius column encodes (radius = beadOuterR/(1+BeadRingTubeRatio)). So spacing
// has NO residual deviation left to bound, unlike the earlier stretched-spacing model this
// test used to pin: a wildly wrong spacing formula (a stray factor, or spacing not tracking
// the per-edge radius) fails this to float-round-off tolerance, not merely "looking odd" on
// screen.
func TestChainBeadsSpacingExactlyTwiceBeadOuterR(t *testing.T) {
	const count = 20
	quantIRForCount := count + nodeTorusSteps("Input") + nodeTorusSteps("Input")
	for _, residueFrac := range []float64{-0.5, -0.2, 0, 0.2, 0.5} {
		actualGap := (float64(count) + residueFrac) * wire.BeadStepR
		m := centerGapNotOnLattice("Input", "Input", quantIRForCount, actualGap)
		ox, oy, oz, _, _, radius := m.chainBeads()
		if len(ox) != count {
			t.Fatalf("residue %.2f: bead count = %d, want %d", residueFrac, len(ox), count)
		}
		for i := 1; i < len(ox); i++ {
			dx, dy, dz := float64(ox[i])-float64(ox[i-1]), float64(oy[i])-float64(oy[i-1]), float64(oz[i])-float64(oz[i-1])
			spacing := math.Sqrt(dx*dx + dy*dy + dz*dz)
			wantSpacing := 2 * beadOuterRFromSphereR(radius[i])
			if math.Abs(spacing-wantSpacing) > tangencyEps {
				t.Errorf("residue %.2f: bead %d..%d spacing %.6f, want exactly 2*beadOuterR %.6f (beads touch on a straight chain)",
					residueFrac, i-1, i, spacing, wantSpacing)
			}
		}
	}
}

// --- "the chain aims where the node is" (docs/bead-lattice.md residue, DIRECTION fix) ---
//
// The tests above never trip the bearing defect either: centerGapNotOnLattice always
// places the live target along the SAME direction its stored LocalPolar bearing already
// implies (QuantITheta=QuantIPhi=0), so a chain aimed by the stale stored cell would still
// land in the right place by coincidence. offAxisFixture below deliberately puts the
// STORED bearing and the LIVE bearing ~53 degrees apart, so a chain aimed by the stored
// cell instead of the live measurement cannot pass by accident.

// offAxisFixture builds a *nodeMover for one edge "a"->"b" whose STORED LocalPolar bearing
// (QuantITheta/QuantIPhi, an exact 1-degree-cell direction) and whose LIVE partnerCenters
// direction are deliberately DIFFERENT directions — the stored bearing points along +Y
// (QuantITheta=0, straight along the pole), the live center sits off to the side in the
// X/Z plane at colatitude ~53.13 degrees (a 3-4-5 triangle's angle, chosen only because
// it is not a whole degree and not a special angle, so no accidental alignment). A chain
// aimed by the stored cell instead of the live measurement lands roughly `actualGap`
// world units away from the target's actual surface point — the reported "about a third"
// of the ORIGINAL (pre both fixes) gap, at this fixture's separations on the order of a
// world unit, not a rounding-noise fraction of one.
func offAxisFixture(srcKind, dstKind string, quantIRForCount int, actualGap float64) *nodeMover {
	lh := &wire.LayoutHolder{}
	// Stored bearing: straight along the pole (colatitude 0). Deliberately NOT the live
	// direction below, so any code path that still reads QuantITheta/QuantIPhi for
	// PLACEMENT (rather than only as edgeStepCount's length input) aims the chain roughly
	// 53 degrees away from where the fixture's live center actually is.
	lh.SetLocalPolar("b", 0, 0, quantIRForCount, 0, 0, 0)
	selfTorus := nodeTorusOuterR(srcKind)
	dstTorus := nodeTorusOuterR(dstKind)
	dist := selfTorus + actualGap + dstTorus
	// Live direction: (3,0,4)/5 — a unit vector well off the stored pole direction, with
	// no sqrt needed to state exactly (a 3-4-5 triangle), scaled to the exact required
	// center-to-center distance.
	targetCenter := vec3{X: dist * 0.6, Y: 0, Z: dist * 0.8}
	return &nodeMover{
		id:             "a",
		geom:           nodeGeom{nodeIdentity: nodeIdentity{Kind: srcKind}}, // HasPos false -> center at origin
		outTargets:     []string{"b"},
		cascadeKinds:   map[string]string{"b": dstKind},
		layoutHolderFn: func() *wire.LayoutHolder { return lh },
		partnerCenters: map[string]vec3{"b": targetCenter},
	}
}

// THE FIX's regression test for BEARING (the length-only fix's TestChainBeadsTangentToLiveGapNotOnLattice
// counterpart): the last bead's centre must sit at the exact 3D distance
// targetTorusR+BeadTorusOuterR from the TARGET's own live centre — not merely at the right
// distance from the origin along SOME axis, which is exactly the assertion a purely
// along-axis check would pass even while the chain points the wrong way (this fixture's
// stored bearing and live bearing are ~53 degrees apart, so an along-axis-only check
// against the WRONG axis would still read as "tangent"). The first bead's near edge from
// the origin (this node's own centre) is checked the same way the other tests do, since
// that end is unaffected by direction (bead 0 sits at the fixed radial distance
// selfTorusR+BeadTorusOuterR from THIS node regardless of which way the chain points).
func TestChainBeadsLastBeadOnTargetTorusOffAxis(t *testing.T) {
	const count = 12
	quantIRForCount := count + nodeTorusSteps("Input") + nodeTorusSteps("Time")
	// A gap deliberately off the 1-degree-cell lattice too, so this test also exercises
	// the length fix at the same time as the bearing fix — the two residues are
	// independent and both must be closed.
	actualGap := (float64(count) + 0.41) * wire.BeadStepR
	m := offAxisFixture("Input", "Time", quantIRForCount, actualGap)
	targetCenter := m.partnerCenters["b"]
	srcTorus := nodeTorusOuterR("Input")
	dstTorus := nodeTorusOuterR("Time")

	ox, oy, oz, _, _, radius := m.chainBeads()
	if len(ox) != count {
		t.Fatalf("bead count = %d, want edgeStepCount's %d", len(ox), count)
	}

	// Bead 0's near edge, from the origin (this node's own live centre) — unaffected by
	// bearing, so this pins the length end of the invariant only.
	d0 := math.Sqrt(float64(ox[0])*float64(ox[0]) + float64(oy[0])*float64(oy[0]) + float64(oz[0])*float64(oz[0]))
	if got, want := d0-beadOuterRFromSphereR(radius[0]), srcTorus; math.Abs(got-want) > tangencyEps {
		t.Errorf("bead 0 near edge %.9f, want exactly source torus %.9f", got, want)
	}

	// Last bead's far edge, measured as a full 3D distance TO THE TARGET'S OWN CENTRE
	// (not the origin, not a single axis) — the assertion that catches a lateral miss.
	last := len(ox) - 1
	dx := float64(ox[last]) - targetCenter.X
	dy := float64(oy[last]) - targetCenter.Y
	dz := float64(oz[last]) - targetCenter.Z
	distToTarget := math.Sqrt(dx*dx + dy*dy + dz*dz)
	wantDist := dstTorus + beadOuterRFromSphereR(radius[last])
	if math.Abs(distToTarget-wantDist) > tangencyEps {
		t.Errorf("last bead (%d) to target centre = %.9f, want exactly targetTorusR+BeadTorusOuterR = %.9f (off by %.9f)",
			last, distToTarget, wantDist, distToTarget-wantDist)
	}
}
