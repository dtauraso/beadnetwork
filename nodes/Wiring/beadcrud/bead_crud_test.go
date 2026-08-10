package beadcrud

import "testing"

// bead_crud_test.go pins beadCrudDecide's invariants directly — PLAN.md's Tests section,
// at the single-bead granularity this function judges. commitNodeMoveLocal's own tests
// (quantized_layout_test.go) cover the multi-edge wiring on top of this.

const testBeadLen = 8.0

// TestBeadCrudUsesSourcePointNotBeadCentre pins the exact bullet PLAN.md calls out: the
// third vector must be built from the bead's SOURCE point, not its own centre — using the
// centre instead is "wrong by one bead" and gives a plausible-looking but wrong verdict.
func TestBeadCrudUsesSourcePointNotBeadCentre(t *testing.T) {
	source := vec3{X: 0, Y: 0, Z: 0}
	centre := vec3{X: testBeadLen, Y: 0, Z: 0} // one bead length from source
	// Destination two bead lengths from source, along the same axis: a real gap has
	// opened (add), and the drag vector points the same direction as the bead vector
	// (0 degrees), so the gate admits it.
	dest := vec3{X: 2 * testBeadLen, Y: 0, Z: 0}
	drag := dest.Sub(source)

	verdict, third := BeadCrudDecide(source, centre, dest, drag, testBeadLen)
	if verdict != BeadCrudAdd {
		t.Fatalf("verdict from the correct source point = %v, want BeadCrudAdd", verdict)
	}
	if got := third.Length(); got != 2*testBeadLen {
		t.Fatalf("third length = %v, want %v (source-to-destination)", got, 2*testBeadLen)
	}

	// Using the bead's own CENTRE as the source (the mistake PLAN.md warns against)
	// measures only ONE bead length of span (dest-centre), not two — a materially
	// different, wrong verdict (none, not add) for the identical drag.
	wrongVerdict, wrongThird := BeadCrudDecide(centre, centre, dest, drag, testBeadLen)
	if wrongVerdict == verdict && wrongThird.Length() == third.Length() {
		t.Fatal("test setup invalid: using the bead's own centre as source must disagree with using the true source point")
	}
	if wrongVerdict != BeadCrudNone {
		t.Fatalf("using the bead's own centre as source gave verdict %v, want the off-by-one-bead answer BeadCrudNone", wrongVerdict)
	}
}

// TestBeadCrudRemoveWhenSpanTooShort: |third| shorter than one bead length removes the
// touching bead — no angle gate involved.
func TestBeadCrudRemoveWhenSpanTooShort(t *testing.T) {
	source := vec3{X: 0, Y: 0, Z: 0}
	centre := vec3{X: testBeadLen, Y: 0, Z: 0}
	dest := vec3{X: testBeadLen * 0.4, Y: 0, Z: 0} // span < one bead length
	drag := dest.Sub(source)

	verdict, _ := BeadCrudDecide(source, centre, dest, drag, testBeadLen)
	if verdict != BeadCrudRemove {
		t.Fatalf("verdict = %v, want BeadCrudRemove", verdict)
	}
}

// TestBeadCrudExactBeadLengthMovesNothing: a drag whose third vector comes out at exactly
// one bead length changes nothing.
func TestBeadCrudExactBeadLengthMovesNothing(t *testing.T) {
	source := vec3{X: 0, Y: 0, Z: 0}
	centre := vec3{X: testBeadLen, Y: 0, Z: 0}
	dest := vec3{X: testBeadLen, Y: 0, Z: 0} // |third| == beadLen exactly
	drag := dest.Sub(source)

	verdict, _ := BeadCrudDecide(source, centre, dest, drag, testBeadLen)
	if verdict != BeadCrudNone {
		t.Fatalf("verdict = %v, want BeadCrudNone", verdict)
	}
}

// TestBeadCrudAngleGateBlocksAddOnly: with the angle between the drag vector and the
// edge-bead vector greater than 90 degrees, no bead is added even though |third| exceeds
// one bead length — the SAME configuration with the angle at or under 90 degrees does add
// one. A removal driven by |third| alone must be unaffected by the angle, so the gate
// cannot be applied to both cases by mistake.
func TestBeadCrudAngleGateBlocksAddOnly(t *testing.T) {
	source := vec3{X: 0, Y: 0, Z: 0}
	centre := vec3{X: testBeadLen, Y: 0, Z: 0}

	// Destination far past centre in the SAME direction as the bead vector (0 degree
	// angle): |third| is large (add-worthy) and the gate admits it.
	forwardDest := vec3{X: 3 * testBeadLen, Y: 0, Z: 0}
	forwardDrag := forwardDest.Sub(source)
	verdict, _ := BeadCrudDecide(source, centre, forwardDest, forwardDrag, testBeadLen)
	if verdict != BeadCrudAdd {
		t.Fatalf("aligned drag: verdict = %v, want BeadCrudAdd", verdict)
	}

	// Destination far in the OPPOSITE direction from source (drag heading back across
	// the bead, not opening a gap beyond it): |third| is still large enough to ask for
	// an add (span source->dest is huge), but the angle between beadVec and the drag
	// vector is 180 degrees, so the gate must block it.
	backwardDest := vec3{X: -3 * testBeadLen, Y: 0, Z: 0}
	backwardDrag := backwardDest.Sub(source)
	blockedVerdict, blockedThird := BeadCrudDecide(source, centre, backwardDest, backwardDrag, testBeadLen)
	if blockedThird.Length() <= testBeadLen {
		t.Fatalf("test setup invalid: backward third length %v must exceed one bead length for the gate to be exercised", blockedThird.Length())
	}
	if blockedVerdict != BeadCrudNone {
		t.Fatalf("backward drag past the gate: verdict = %v, want BeadCrudNone (angle gate must block the add)", blockedVerdict)
	}

	// A REMOVAL is never gated: drive |third| below one bead length along the SAME
	// "backward" heading the add gate just blocked, and it must still remove.
	removeDest := vec3{X: testBeadLen * 0.2, Y: 0, Z: 0}
	removeDrag := removeDest.Sub(source)
	removeVerdict, removeThird := BeadCrudDecide(source, centre, removeDest, removeDrag, testBeadLen)
	if removeThird.Length() >= testBeadLen {
		t.Fatalf("test setup invalid: remove third length %v must be under one bead length", removeThird.Length())
	}
	if removeVerdict != BeadCrudRemove {
		t.Fatalf("verdict = %v, want BeadCrudRemove — a removal must not be affected by the angle gate", removeVerdict)
	}
}

// TestBeadCrudNoConfigurationDependence: the verdict is a pure function of the three
// points, the drag vector, and beadLen — it reads no radius, colatitude, or
// neighbour-set shape. Two unrelated configurations (different radii/positions entirely)
// that happen to produce the same beadSource/beadVec/third/drag relationship must produce
// the same verdict.
func TestBeadCrudNoConfigurationDependence(t *testing.T) {
	// Configuration A: small radius, axis-aligned.
	sourceA := vec3{X: 0, Y: 0, Z: 0}
	centreA := vec3{X: testBeadLen, Y: 0, Z: 0}
	destA := vec3{X: 2 * testBeadLen, Y: 0, Z: 0}
	dragA := destA.Sub(sourceA)
	verdictA, _ := BeadCrudDecide(sourceA, centreA, destA, dragA, testBeadLen)

	// Configuration B: same relative geometry, translated and rotated into a totally
	// different part of space (large radius / different colatitude), same bead length.
	offset := vec3{X: 500, Y: -300, Z: 900}
	// Rotate 90 degrees about Z: (x,y,z) -> (-y,x,z).
	rot := func(v vec3) vec3 { return vec3{X: -v.Y, Y: v.X, Z: v.Z} }
	sourceB := rot(sourceA).Add(offset)
	centreB := rot(centreA).Add(offset)
	destB := rot(destA).Add(offset)
	dragB := rot(dragA)
	verdictB, _ := BeadCrudDecide(sourceB, centreB, destB, dragB, testBeadLen)

	if verdictA != verdictB {
		t.Fatalf("verdict depended on configuration: A=%v B=%v for the same relative geometry", verdictA, verdictB)
	}
}

// TestResolveBeadCrudMoveThreeNeighborsDisagreeStillMoves pins the exact bug that shipped:
// a node with THREE neighbours has touching beads on three different chain axes, so their
// implied centres essentially never coincide — this is the ORDINARY multi-neighbour case,
// not a conflict, and the node must still MOVE (never hold position). An earlier version
// treated any disagreement among implied centres as a conflict and held the node still,
// which made every multi-neighbour node immovable (node 2 could not be dragged at all).
// resolveBeadCrudMove must pick the implied centre with the SMALLEST displacement from
// prevPos — never an average of the three, never the raw drag/mouse target.
func TestResolveBeadCrudMoveThreeNeighborsDisagreeStillMoves(t *testing.T) {
	prevPos := vec3{X: 0, Y: 0, Z: 0}
	nodeDestination := vec3{X: 1, Y: 0, Z: 0}

	// Three touching beads, one per neighbour, on three different chain axes (X, Y, Z).
	// Each beadSource sits close to nodeDestination so every one judges REMOVE
	// independently — REMOVE's implied centre is the bead's own current centre exactly
	// (beadCrudImpliedCentre), so the three implied centres below (X=5, Y=3, Z=7) are
	// necessarily different points, by construction, not by accident.
	beads := []TouchingBead{
		{NeighborID: "N1", Source: vec3{X: 1, Y: 0, Z: 0}, Centre: vec3{X: 5, Y: 0, Z: 0}, AimDir: vec3{X: 1, Y: 0, Z: 0}},
		{NeighborID: "N2", Source: vec3{X: 0, Y: 1, Z: 0}, Centre: vec3{X: 0, Y: 3, Z: 0}, AimDir: vec3{X: 0, Y: 1, Z: 0}},
		{NeighborID: "N3", Source: vec3{X: 0, Y: 0, Z: 1}, Centre: vec3{X: 0, Y: 0, Z: 7}, AimDir: vec3{X: 0, Y: 0, Z: 1}},
	}

	committed, results := ResolveBeadCrudMove(beads, prevPos, nodeDestination, testBeadLen)

	if len(results) != 3 {
		t.Fatalf("expected all three touching beads to reach a non-none verdict, got %d results: %+v", len(results), results)
	}
	for _, r := range results {
		if r.Verdict != BeadCrudRemove {
			t.Fatalf("test setup invalid: expected every bead to judge REMOVE, got %v for %s", r.Verdict, r.NeighborID)
		}
	}

	// The node must MOVE — never hold prevPos on disagreement.
	if committed == prevPos {
		t.Fatal("node held its previous position on a three-way disagreement — this is the shipped bug: a multi-neighbour node must still move")
	}

	// The commit must be one of the per-bead implied centres exactly (N2's, the smallest
	// displacement from prevPos: |3| < |5| < |7|) — never an average of the three, and
	// never the raw drag/mouse target.
	want := vec3{X: 0, Y: 3, Z: 0}
	if committed != want {
		t.Fatalf("committed = %+v, want %+v (N2's implied centre, the smallest displacement from prevPos)", committed, want)
	}

	avg := vec3{
		X: (5 + 0 + 0) / 3,
		Y: (0 + 3 + 0) / 3,
		Z: (0 + 0 + 7) / 3,
	}
	if committed == avg {
		t.Fatal("committed equals the average of the three implied centres — averaging is forbidden")
	}
	if committed == nodeDestination {
		t.Fatal("committed equals the raw mouse/drag target — reading the mouse target is forbidden")
	}

	// The commit must exactly match one of the results' own Implied centre — never a
	// synthesized point.
	matched := false
	for _, r := range results {
		if r.Implied == committed {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("committed %+v does not match any per-bead implied centre in results %+v", committed, results)
	}
}
