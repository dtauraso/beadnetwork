package Wiring

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

	verdict, third := beadCrudDecide(source, centre, dest, drag, testBeadLen)
	if verdict != beadCrudAdd {
		t.Fatalf("verdict from the correct source point = %v, want beadCrudAdd", verdict)
	}
	if got := third.Length(); got != 2*testBeadLen {
		t.Fatalf("third length = %v, want %v (source-to-destination)", got, 2*testBeadLen)
	}

	// Using the bead's own CENTRE as the source (the mistake PLAN.md warns against)
	// measures only ONE bead length of span (dest-centre), not two — a materially
	// different, wrong verdict (none, not add) for the identical drag.
	wrongVerdict, wrongThird := beadCrudDecide(centre, centre, dest, drag, testBeadLen)
	if wrongVerdict == verdict && wrongThird.Length() == third.Length() {
		t.Fatal("test setup invalid: using the bead's own centre as source must disagree with using the true source point")
	}
	if wrongVerdict != beadCrudNone {
		t.Fatalf("using the bead's own centre as source gave verdict %v, want the off-by-one-bead answer beadCrudNone", wrongVerdict)
	}
}

// TestBeadCrudRemoveWhenSpanTooShort: |third| shorter than one bead length removes the
// touching bead — no angle gate involved.
func TestBeadCrudRemoveWhenSpanTooShort(t *testing.T) {
	source := vec3{X: 0, Y: 0, Z: 0}
	centre := vec3{X: testBeadLen, Y: 0, Z: 0}
	dest := vec3{X: testBeadLen * 0.4, Y: 0, Z: 0} // span < one bead length
	drag := dest.Sub(source)

	verdict, _ := beadCrudDecide(source, centre, dest, drag, testBeadLen)
	if verdict != beadCrudRemove {
		t.Fatalf("verdict = %v, want beadCrudRemove", verdict)
	}
}

// TestBeadCrudExactBeadLengthMovesNothing: a drag whose third vector comes out at exactly
// one bead length changes nothing.
func TestBeadCrudExactBeadLengthMovesNothing(t *testing.T) {
	source := vec3{X: 0, Y: 0, Z: 0}
	centre := vec3{X: testBeadLen, Y: 0, Z: 0}
	dest := vec3{X: testBeadLen, Y: 0, Z: 0} // |third| == beadLen exactly
	drag := dest.Sub(source)

	verdict, _ := beadCrudDecide(source, centre, dest, drag, testBeadLen)
	if verdict != beadCrudNone {
		t.Fatalf("verdict = %v, want beadCrudNone", verdict)
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
	verdict, _ := beadCrudDecide(source, centre, forwardDest, forwardDrag, testBeadLen)
	if verdict != beadCrudAdd {
		t.Fatalf("aligned drag: verdict = %v, want beadCrudAdd", verdict)
	}

	// Destination far in the OPPOSITE direction from source (drag heading back across
	// the bead, not opening a gap beyond it): |third| is still large enough to ask for
	// an add (span source->dest is huge), but the angle between beadVec and the drag
	// vector is 180 degrees, so the gate must block it.
	backwardDest := vec3{X: -3 * testBeadLen, Y: 0, Z: 0}
	backwardDrag := backwardDest.Sub(source)
	blockedVerdict, blockedThird := beadCrudDecide(source, centre, backwardDest, backwardDrag, testBeadLen)
	if blockedThird.Length() <= testBeadLen {
		t.Fatalf("test setup invalid: backward third length %v must exceed one bead length for the gate to be exercised", blockedThird.Length())
	}
	if blockedVerdict != beadCrudNone {
		t.Fatalf("backward drag past the gate: verdict = %v, want beadCrudNone (angle gate must block the add)", blockedVerdict)
	}

	// A REMOVAL is never gated: drive |third| below one bead length along the SAME
	// "backward" heading the add gate just blocked, and it must still remove.
	removeDest := vec3{X: testBeadLen * 0.2, Y: 0, Z: 0}
	removeDrag := removeDest.Sub(source)
	removeVerdict, removeThird := beadCrudDecide(source, centre, removeDest, removeDrag, testBeadLen)
	if removeThird.Length() >= testBeadLen {
		t.Fatalf("test setup invalid: remove third length %v must be under one bead length", removeThird.Length())
	}
	if removeVerdict != beadCrudRemove {
		t.Fatalf("verdict = %v, want beadCrudRemove — a removal must not be affected by the angle gate", removeVerdict)
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
	verdictA, _ := beadCrudDecide(sourceA, centreA, destA, dragA, testBeadLen)

	// Configuration B: same relative geometry, translated and rotated into a totally
	// different part of space (large radius / different colatitude), same bead length.
	offset := vec3{X: 500, Y: -300, Z: 900}
	// Rotate 90 degrees about Z: (x,y,z) -> (-y,x,z).
	rot := func(v vec3) vec3 { return vec3{X: -v.Y, Y: v.X, Z: v.Z} }
	sourceB := rot(sourceA).Add(offset)
	centreB := rot(centreA).Add(offset)
	destB := rot(destA).Add(offset)
	dragB := rot(dragA)
	verdictB, _ := beadCrudDecide(sourceB, centreB, destB, dragB, testBeadLen)

	if verdictA != verdictB {
		t.Fatalf("verdict depended on configuration: A=%v B=%v for the same relative geometry", verdictA, verdictB)
	}
}
