package nodeactor

import (
	"math"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// chain_beads_placement_test.go — chainBeads' placement invariants: beads never sit inside
// either node's sphere, adjacent beads touch, a collapsed edge still gets one bead, an
// unknown partner contributes nothing, count scales with span, and the exact double
// tangency at both ends (on axis and off it).

// The invariant the original tests missed, and the bug that shipped: a bead must never sit
// inside either node's sphere — restated in TANGENCY terms under the bead lattice
// (docs/bead-model/bead-lattice.md "Placement"): bead 0's torus is tangent OUTSIDE the source node's
// torus, and the last bead's torus is tangent OUTSIDE the target's, never overlapping either.
func TestChainBeadsStayOutsideBothNodes(t *testing.T) {
	// Expressed as a cell count * lattice.BeadStepR, not a bare literal, so this
	// stays an exact multiple of the local-polar grid constant whatever that constant is.
	const gap = 200 * lattice.BeadStepR
	m := &NodeGeometry{
		id:   "a",
		geom: nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: "Input"}}, // radius 15
		outs: nodeOuts{outTargets: []string{"b"}},
		topo: neighborTopology{
			neighborKinds:  map[string]string{"b": "Time"}, // radius 9
			partnerCenters: singleNeighborCenter("b", gap),
		},
	}
	ox, oy, oz, _, _, _ := m.chainBeads()
	if len(ox) == 0 {
		t.Fatal("no beads emitted for a 400-unit gap")
	}
	srcClear := nodegeom.NodeTorusOuterR("Input")
	dstClear := gap - nodegeom.NodeTorusOuterR("Time")
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
// (lattice.BeadStepR), so a chain is a solid line with no gaps. Spacing is the single
// UNIFORM lattice constant now (no per-edge sizing — this fixture's neighbor is placed at
// an exact multiple of lattice.BeadStepR, so there is no residue for a per-edge size to
// absorb; MODEL.md "Moving a node is CRUD on the edge beads that touch it").
func TestChainBeadsTouch(t *testing.T) {
	m := &NodeGeometry{
		id: "a", geom: nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: "Input"}},
		outs: nodeOuts{outTargets: []string{"b"}},
		topo: neighborTopology{
			neighborKinds: map[string]string{"b": "Input"},
			// 150 cells * lattice.BeadStepR, not the bare literal 300 — see the comment
			// on TestChainBeadsStayOutsideBothNodes's gap.
			partnerCenters: singleNeighborCenter("b", 150*lattice.BeadStepR),
		},
	}
	ox, oy, oz, _, _, _ := m.chainBeads()
	if len(ox) < 3 {
		t.Fatalf("want several beads to compare spacing, got %d", len(ox))
	}
	want := lattice.BeadStepR
	for i := 1; i < len(ox); i++ {
		dx, dy, dz := ox[i]-ox[i-1], oy[i]-oy[i-1], oz[i]-oz[i-1]
		got := math.Sqrt(float64(dx*dx + dy*dy + dz*dz))
		if math.Abs(got-want) > 1e-3 {
			t.Errorf("beads %d..%d are %.4f apart, want one bead step %.4f (touching, no gap)", i-1, i, got, want)
		}
	}
}

// Two nodes whose TORI are closer than one bead step apart still contribute at least
// ONE bead — edgeStepCount clamps to a minimum of 1 (docs/bead-model/bead-lattice.md "The count") so an
// edge never has zero beads, even when the gap collapses.
func TestChainBeadsAlwaysAtLeastOneBead(t *testing.T) {
	m := &NodeGeometry{
		id: "a", geom: nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: "Input"}},
		outs: nodeOuts{outTargets: []string{"b"}},
		topo: neighborTopology{
			neighborKinds: map[string]string{"b": "Input"},
			// 3 cells * lattice.BeadStepR, not the bare literal 6 — see the comment on
			// TestChainBeadsStayOutsideBothNodes's gap. QuantIR IS the bead-step count now
			// (edgeStepCount no longer divides by a cell-per-step constant — bead_lattice.go's
			// BeadStepCells doc comment), so 3 steps minus both "Input" nodes' own
			// nodeTorusSteps (2 each) collapses well below the minimum.
			partnerCenters: singleNeighborCenter("b", 3*lattice.BeadStepR),
		},
	}
	if ox, _, _, _, _, _ := m.chainBeads(); len(ox) != 1 {
		t.Errorf("count = %d, want 1 — edgeStepCount clamps a collapsed gap to the minimum, never 0", len(ox))
	}
}

// A target this node has no live partner center for contributes nothing — the node aims
// only with its own live m.partnerCenters, never a stored bearing or a made-up direction
// (MODEL.md "the polar model": no node-node stored coordinate).
func TestChainBeadsUnknownPartnerContributesNothing(t *testing.T) {
	m := &NodeGeometry{
		id: "a", geom: nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: "Input"}},
		outs: nodeOuts{outTargets: []string{"b"}},
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
		m := &NodeGeometry{
			id: "a", geom: nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: "Input"}},
			outs: nodeOuts{outTargets: []string{"b"}},
			topo: neighborTopology{
				neighborKinds:  map[string]string{"b": "Input"},
				partnerCenters: singleNeighborCenter("b", centerGap),
			},
		}
		ox, _, _, _, _, _ := m.chainBeads()
		return len(ox)
	}
	// base and span must each be an exact multiple of lattice.BeadStepR
	// (singleNeighborHolder's requirement — a sum of two such multiples is one too, so
	// base+span and base+2*span stay exact). base is chosen so it exactly cancels the
	// two nodeTorusSteps("Input") subtractions (2 each, 4 total, edgeStepCount): with
	// QuantIR itself the bead-step count (no per-step division anymore — bead_lattice.go's
	// BeadStepCells doc comment), count(base+span) = base+span-4 = span when base=4, and
	// count(base+2*span) = 2*span exactly — an exact double, not merely "roughly", because
	// there is nothing left to round.
	const base = 4 * lattice.BeadStepR
	span := 80 * lattice.BeadStepR
	n1 := count(base + span)
	n2 := count(base + 2*span)
	if n2 != 2*n1 {
		t.Errorf("count(span %.0f) = %d, count(span %.0f) = %d; want exactly double", span, n1, 2*span, n2)
	}
}

// THE REGRESSION GUARD for this commit: exact double tangency, no float tolerance wider
// than round-trip noise. Before this commit, edgeStepCount measured
// `round((QuantIR*stepR - nodegeom.NodeTorusOuterR(src) - nodegeom.NodeTorusOuterR(dst)) / BeadStepR)` against
// an nodegeom.NodeTorusOuterR that was an arbitrary float NOT on the bead lattice
// (nodegeom.NodeRadius(kind)*(1+ShadingParamNodeRingTubeRatio)), so the division was essentially never
// exact and the rounding silently absorbed up to half a bead step at the TARGET end — bead 0
// was always exactly tangent to the source (offset by construction), but the last bead's far
// edge only APPROXIMATELY met the target's torus. This test pins the far edge to the target's
// torus to float-round-off tolerance (1e-3, chainBeads' streamed float32 offsets — see
// tangencyEps's own comment), across several distance values this fixture deliberately
// snaps onto the bead lattice (singleNeighborHolder) and several node-kind pairs whose
// bareNodeRadius values don't share an obvious common factor.
func TestChainBeadsExactDoubleTangency(t *testing.T) {
	kindPairs := [][2]string{
		{"Input", "Time"},
		{"Time", "Input"},
		{"Input", "Input"},
		{"Time", "Time"},
	}
	// Every centerGap here must be an exact multiple of lattice.BeadStepR
	// (singleNeighborHolder's own requirement) and large enough to clear both tori with
	// room for at least one bead, across the largest kind pair tested.
	cellCounts := []float64{100, 120, 180, 260, 500}
	centerGaps := make([]float64, len(cellCounts))
	for i, c := range cellCounts {
		centerGaps[i] = c * lattice.BeadStepR
	}

	for _, kp := range kindPairs {
		srcKind, dstKind := kp[0], kp[1]
		for _, gap := range centerGaps {
			m := &NodeGeometry{
				id:   "a",
				geom: nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: srcKind}},
				outs: nodeOuts{outTargets: []string{"b"}},
				topo: neighborTopology{
					neighborKinds:  map[string]string{"b": dstKind},
					partnerCenters: singleNeighborCenter("b", gap),
				},
			}
			ox, oy, oz, _, _, _ := m.chainBeads()
			if len(ox) == 0 {
				t.Fatalf("%s->%s gap %.0f: no beads emitted", srcKind, dstKind, gap)
			}

			srcTorus := nodegeom.NodeTorusOuterR(srcKind)
			dstTorus := nodegeom.NodeTorusOuterR(dstKind)

			// Bead 0's NEAR edge: its offset from center minus its own torus radius must
			// equal the source's torus radius EXACTLY.
			d0 := math.Sqrt(float64(ox[0])*float64(ox[0]) + float64(oy[0])*float64(oy[0]) + float64(oz[0])*float64(oz[0]))
			nearEdge0 := d0 - lattice.BeadTorusOuterR
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
			farEdgeLast := dLast + lattice.BeadTorusOuterR
			wantFarEdge := gap - dstTorus
			if math.Abs(farEdgeLast-wantFarEdge) > tangencyEps {
				t.Errorf("%s->%s gap %.0f: last bead (%d) far edge %.9f, want exactly target's near torus edge %.9f (off by %.9f, up to half a bead step %.3f would indicate math.Round is back)",
					srcKind, dstKind, gap, last, farEdgeLast, wantFarEdge, farEdgeLast-wantFarEdge, lattice.BeadStepR/2)
			}
		}
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
	targetCenter := m.topo.partnerCenters["b"]
	srcTorus := nodegeom.NodeTorusOuterR("Input")
	dstTorus := nodegeom.NodeTorusOuterR("Time")

	ox, oy, oz, _, _, _ := m.chainBeads()
	if len(ox) != count {
		t.Fatalf("bead count = %d, want edgeStepCount's %d", len(ox), count)
	}

	// Bead 0's near edge, from the origin (this node's own live centre) — unaffected by
	// bearing, so this pins the length end of the invariant only.
	d0 := math.Sqrt(float64(ox[0])*float64(ox[0]) + float64(oy[0])*float64(oy[0]) + float64(oz[0])*float64(oz[0]))
	if got, want := d0-lattice.BeadTorusOuterR, srcTorus; math.Abs(got-want) > tangencyEps {
		t.Errorf("bead 0 near edge %.9f, want exactly source torus %.9f", got, want)
	}

	// Last bead's far edge, measured as a full 3D distance TO THE TARGET'S OWN CENTRE
	// (not the origin, not a single axis) — the assertion that catches a lateral miss.
	last := len(ox) - 1
	dx := float64(ox[last]) - targetCenter.X
	dy := float64(oy[last]) - targetCenter.Y
	dz := float64(oz[last]) - targetCenter.Z
	distToTarget := math.Sqrt(dx*dx + dy*dy + dz*dz)
	wantDist := dstTorus + lattice.BeadTorusOuterR
	if math.Abs(distToTarget-wantDist) > tangencyEps {
		t.Errorf("last bead (%d) to target centre = %.9f, want exactly targetTorusR+BeadTorusOuterR = %.9f (off by %.9f)",
			last, distToTarget, wantDist, distToTarget-wantDist)
	}
}
