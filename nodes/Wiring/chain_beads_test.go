package Wiring

import (
	"math"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// chainBeads is a pure function of ONE node's own state — its own kind, its own
// neighborKinds, and its own live copy of the neighbour's world center
// (m.partnerCenters) — all written only by that node's goroutine — so these are plain
// tables, no second goroutine (docs/testing-shape.md).
//
// Every case sets BOTH kinds explicitly. An unset kind is NOT neutral: kindWidthHeight falls
// back to (110, 60), i.e. radius 15. The original tests left kinds unset and asserted on
// distance from the node CENTER, which is exactly why they passed while beads rendered inside
// the nodes.
//
// A neighbour's position is supplied the same way production reads it now: a live cartesian
// center in m.partnerCenters, placed straight along +Y at the given center-to-center
// distance (an arbitrary but fixed direction; magnitude-only assertions below don't care
// which one). MODEL.md "the polar model": there is no stored node-node bearing any more
// (wire.LocalPolar deleted) — a node's aim to an edge's first bead is measured live.

// singleNeighborCenter builds a partnerCenters map with exactly one entry — `to` placed
// centerGap world units along +Y from the origin (this node's own center, since a bare
// nodeMover literal has HasPos false).
func singleNeighborCenter(to string, centerGap float64) map[string]vec3 {
	return map[string]vec3{to: {X: 0, Y: centerGap, Z: 0}}
}

// The invariant the original tests missed, and the bug that shipped: a bead must never sit
// inside either node's sphere — restated in TANGENCY terms under the bead lattice
// (docs/bead-lattice.md "Placement"): bead 0's torus is tangent OUTSIDE the source node's
// torus, and the last bead's torus is tangent OUTSIDE the target's, never overlapping either.
func TestChainBeadsStayOutsideBothNodes(t *testing.T) {
	// Expressed as a cell count * wire.BeadStepR, not a bare literal, so this
	// stays an exact multiple of the local-polar grid constant whatever that constant is.
	const gap = 200 * wire.BeadStepR
	m := &nodeGeometry{
		id:             "a",
		geom:           nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}}, // radius 15
		outTargets:     []string{"b"},
		neighborKinds:  map[string]string{"b": "Time"}, // radius 9
		partnerCenters: singleNeighborCenter("b", gap),
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
// (wire.BeadStepR), so a chain is a solid line with no gaps. Spacing is the single
// UNIFORM lattice constant now (no per-edge sizing — this fixture's neighbor is placed at
// an exact multiple of wire.BeadStepR, so there is no residue for a per-edge size to
// absorb; MODEL.md "Moving a node is CRUD on the edge beads that touch it").
func TestChainBeadsTouch(t *testing.T) {
	m := &nodeGeometry{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
		outTargets: []string{"b"}, neighborKinds: map[string]string{"b": "Input"},
		// 150 cells * wire.BeadStepR, not the bare literal 300 — see the comment
		// on TestChainBeadsStayOutsideBothNodes's gap.
		partnerCenters: singleNeighborCenter("b", 150*wire.BeadStepR),
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
	m := &nodeGeometry{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
		outTargets: []string{"b"}, neighborKinds: map[string]string{"b": "Input"},
		// 3 cells * wire.BeadStepR, not the bare literal 6 — see the comment on
		// TestChainBeadsStayOutsideBothNodes's gap. QuantIR IS the bead-step count now
		// (edgeStepCount no longer divides by a cell-per-step constant — bead_lattice.go's
		// BeadStepCells doc comment), so 3 steps minus both "Input" nodes' own
		// nodeTorusSteps (2 each) collapses well below the minimum.
		partnerCenters: singleNeighborCenter("b", 3*wire.BeadStepR),
	}
	if ox, _, _, _, _, _ := m.chainBeads(); len(ox) != 1 {
		t.Errorf("count = %d, want 1 — edgeStepCount clamps a collapsed gap to the minimum, never 0", len(ox))
	}
}

// A target this node has no live partner center for contributes nothing — the node aims
// only with its own live m.partnerCenters, never a stored bearing or a made-up direction
// (MODEL.md "the polar model": no node-node stored coordinate).
func TestChainBeadsUnknownPartnerContributesNothing(t *testing.T) {
	m := &nodeGeometry{
		id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}}, outTargets: []string{"b"},
	}
	if ox, _, _, _, _, _ := m.chainBeads(); len(ox) != 0 {
		t.Errorf("count = %d, want 0 for an unknown partner", len(ox))
	}
}

// Count is proportional to the GAP (torus-to-torus, edgeStepCount's `gap`) — not to the
// center distance, which is what the pre-bead-lattice version asserted and what let the
// buried-bead bug through. Double the span, double the beads (±1 for the round), which is
// what makes a constant per-bead dwell a constant visible speed.
func TestChainBeadsCountIsSpanProportional(t *testing.T) {
	count := func(centerGap float64) int {
		m := &nodeGeometry{
			id: "a", geom: nodeGeom{nodeIdentity: nodeIdentity{Kind: "Input"}},
			outTargets: []string{"b"}, neighborKinds: map[string]string{"b": "Input"},
			partnerCenters: singleNeighborCenter("b", centerGap),
		}
		ox, _, _, _, _, _ := m.chainBeads()
		return len(ox)
	}
	// base and span must each be an exact multiple of wire.BeadStepR
	// (singleNeighborHolder's requirement — a sum of two such multiples is one too, so
	// base+span and base+2*span stay exact). base is chosen so it exactly cancels the
	// two nodeTorusSteps("Input") subtractions (2 each, 4 total, edgeStepCount): with
	// QuantIR itself the bead-step count (no per-step division anymore — bead_lattice.go's
	// BeadStepCells doc comment), count(base+span) = base+span-4 = span when base=4, and
	// count(base+2*span) = 2*span exactly — an exact double, not merely "roughly", because
	// there is nothing left to round.
	const base = 4 * wire.BeadStepR
	span := 80 * wire.BeadStepR
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
// chainBeads' node-mover plumbing. It now takes the LIVE center-to-center distance directly
// (K = round(dist/BeadStepR)) rather than a stored LocalPolar, so a distance already an exact
// multiple of BeadStepR (as every test fixture here uses) is plain integer subtraction with
// nothing to round.
func TestEdgeStepCount(t *testing.T) {
	dist := 200 * wire.BeadStepR
	got := edgeStepCount(dist, "Input", "Time")
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
	dist := 1 * wire.BeadStepR // 1 bead step of separation, far inside both tori
	if got := edgeStepCount(dist, "Input", "Time"); got != 1 {
		t.Fatalf("edgeStepCount(collapsed) = %d, want 1", got)
	}
}

// edgeStepCount rounds a NEAR-integer distance to the nearest bead step rather than
// truncating or requiring exactness — the round() exists so a node mid-way through
// placement (a live distance a hair off an exact multiple, from float accumulation) never
// silently drops into the wrong bucket. A node whose live distance happens to land on the
// bead lattice never actually exercises the rounding in practice, but the function must
// still behave sanely on the inputs it can receive.
func TestEdgeStepCountRoundsNearIntegerDistance(t *testing.T) {
	exact := 50 * wire.BeadStepR
	nudged := exact + 1e-6
	if got, want := edgeStepCount(nudged, "Input", "Input"), edgeStepCount(exact, "Input", "Input"); got != want {
		t.Fatalf("edgeStepCount should round a near-integer distance the same as the exact one: got %d want %d", got, want)
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
// torus to float-round-off tolerance (1e-3, chainBeads' streamed float32 offsets — see
// tangencyEps's own comment), across several distance values this fixture deliberately
// snaps onto the bead lattice (singleNeighborHolder) and several node-kind pairs whose
// bareNodeRadius values don't share an obvious common factor.
const tangencyEps = 1e-3

func TestChainBeadsExactDoubleTangency(t *testing.T) {
	kindPairs := [][2]string{
		{"Input", "Time"},
		{"Time", "Input"},
		{"Input", "Input"},
		{"Time", "Time"},
	}
	// Every centerGap here must be an exact multiple of wire.BeadStepR
	// (singleNeighborHolder's own requirement) and large enough to clear both tori with
	// room for at least one bead, across the largest kind pair tested.
	cellCounts := []float64{100, 120, 180, 260, 500}
	centerGaps := make([]float64, len(cellCounts))
	for i, c := range cellCounts {
		centerGaps[i] = c * wire.BeadStepR
	}

	for _, kp := range kindPairs {
		srcKind, dstKind := kp[0], kp[1]
		for _, gap := range centerGaps {
			m := &nodeGeometry{
				id:             "a",
				geom:           nodeGeom{nodeIdentity: nodeIdentity{Kind: srcKind}},
				outTargets:     []string{"b"},
				neighborKinds:  map[string]string{"b": dstKind},
				partnerCenters: singleNeighborCenter("b", gap),
			}
			ox, oy, oz, _, _, _ := m.chainBeads()
			if len(ox) == 0 {
				t.Fatalf("%s->%s gap %.0f: no beads emitted", srcKind, dstKind, gap)
			}

			srcTorus := nodeTorusOuterR(srcKind)
			dstTorus := nodeTorusOuterR(dstKind)

			// Bead 0's NEAR edge: its offset from center minus its own torus radius must
			// equal the source's torus radius EXACTLY.
			d0 := math.Sqrt(float64(ox[0])*float64(ox[0]) + float64(oy[0])*float64(oy[0]) + float64(oz[0])*float64(oz[0]))
			nearEdge0 := d0 - wire.BeadTorusOuterR
			if math.Abs(nearEdge0-srcTorus) > tangencyEps {
				t.Errorf("%s->%s gap %.0f: bead 0 near edge %.9f, want exactly srcTorus %.9f",
					srcKind, dstKind, gap, nearEdge0, srcTorus)
			}

			// Last bead's FAR edge: its offset from center plus its own torus radius must
			// equal the target's torus's NEAR edge (separation - dstTorus) exactly — no
			// residue left once the gap is on the bead lattice, which is now the norm
			// (node placement guarantees it), not a special case.
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

// --- direction aims at the LIVE partner center, independent of spacing/size ---
//
// offAxisFixture places the live partner center off any coordinate axis, so a chain
// placed along the wrong axis (or a made-up direction) cannot pass this test by accident.
// Spacing/size stay the fixed uniform lattice constants regardless (no per-edge sizing
// any more) — this test is purely about DIRECTION. There is no stored bearing any more to
// diverge from (MODEL.md "the polar model": no node-node stored coordinate) — the aim is
// ALWAYS this live measurement.

// offAxisFixture builds a *nodeGeometry for one edge "a"->"b" whose LIVE partnerCenters
// direction sits off to the side in the X/Z plane at colatitude ~53.13 degrees (a 3-4-5
// triangle's angle, chosen only because it is not a whole degree and not a special angle,
// so no accidental alignment). The live center is placed at EXACTLY count*BeadStepR + both
// tori (on-lattice, by this fixture's own construction) so the far-edge tangency assertion
// below has no residue to tolerate beyond float round-off.
func offAxisFixture(srcKind, dstKind string, count int) *nodeGeometry {
	selfTorus := nodeTorusOuterR(srcKind)
	dstTorus := nodeTorusOuterR(dstKind)
	dist := selfTorus + float64(count)*wire.BeadStepR + dstTorus
	// Live direction: (3,0,4)/5 — a unit vector off any coordinate axis, with no sqrt
	// needed to state exactly (a 3-4-5 triangle), scaled to the exact required
	// center-to-center distance.
	targetCenter := vec3{X: dist * 0.6, Y: 0, Z: dist * 0.8}
	return &nodeGeometry{
		id:             "a",
		geom:           nodeGeom{nodeIdentity: nodeIdentity{Kind: srcKind}}, // HasPos false -> center at origin
		outTargets:     []string{"b"},
		neighborKinds:  map[string]string{"b": dstKind},
		partnerCenters: map[string]vec3{"b": targetCenter},
	}
}

// THE FIX's regression test for BEARING: the last bead's centre must sit at the exact 3D
// distance targetTorusR+BeadTorusOuterR from the TARGET's own live centre — not merely at
// the right distance from the origin along SOME axis, which is exactly the assertion a
// purely along-axis check would pass even while the chain points the wrong way (this
// fixture's stored bearing and live bearing are ~53 degrees apart, so an along-axis-only
// check against the WRONG axis would still read as "tangent"). The first bead's near edge
// from the origin (this node's own centre) is checked the same way the other tests do,
// since that end is unaffected by direction.
func TestChainBeadsLastBeadOnTargetTorusOffAxis(t *testing.T) {
	const count = 12
	m := offAxisFixture("Input", "Time", count)
	targetCenter := m.partnerCenters["b"]
	srcTorus := nodeTorusOuterR("Input")
	dstTorus := nodeTorusOuterR("Time")

	ox, oy, oz, _, _, _ := m.chainBeads()
	if len(ox) != count {
		t.Fatalf("bead count = %d, want edgeStepCount's %d", len(ox), count)
	}

	// Bead 0's near edge, from the origin (this node's own live centre) — unaffected by
	// bearing, so this pins the length end of the invariant only.
	d0 := math.Sqrt(float64(ox[0])*float64(ox[0]) + float64(oy[0])*float64(oy[0]) + float64(oz[0])*float64(oz[0]))
	if got, want := d0-wire.BeadTorusOuterR, srcTorus; math.Abs(got-want) > tangencyEps {
		t.Errorf("bead 0 near edge %.9f, want exactly source torus %.9f", got, want)
	}

	// Last bead's far edge, measured as a full 3D distance TO THE TARGET'S OWN CENTRE
	// (not the origin, not a single axis) — the assertion that catches a lateral miss.
	last := len(ox) - 1
	dx := float64(ox[last]) - targetCenter.X
	dy := float64(oy[last]) - targetCenter.Y
	dz := float64(oz[last]) - targetCenter.Z
	distToTarget := math.Sqrt(dx*dx + dy*dy + dz*dz)
	wantDist := dstTorus + wire.BeadTorusOuterR
	if math.Abs(distToTarget-wantDist) > tangencyEps {
		t.Errorf("last bead (%d) to target centre = %.9f, want exactly targetTorusR+BeadTorusOuterR = %.9f (off by %.9f)",
			last, distToTarget, wantDist, distToTarget-wantDist)
	}
}
