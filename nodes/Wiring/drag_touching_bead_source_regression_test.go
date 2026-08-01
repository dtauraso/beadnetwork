package Wiring

// drag_touching_bead_source_regression_test.go — regression coverage for the
// dragTouchingBeads `isSource` bug fixed in quantized_move.go: the SOURCE-side branch used
// to compute `beadSource = prevPos + aimDir*selfTorusR`, a point on THIS node's own torus
// surface — neither of the two source points MODEL.md's bead-CRUD section allows (the
// previous bead's centre along the chain, or the chain origin on the NEIGHBOUR's torus when
// it is the only bead). That broke two things nothing here previously caught:
//
//   - |third| at rest (nodeDestination == the touching bead's own current centre) came out
//     as ~selfTorusR (several bead lengths) instead of one bead length (wire.BeadStepR), so
//     REMOVE (`|third| < beadLen`) could never fire on a source-side edge; and
//   - beadVec (beadCentre - beadSource) pointed TOWARD the neighbour instead of away from
//     it, inverting the ADD angle gate: dragging AWAY from the neighbour (which should open
//     a gap and ADD) scored an obtuse angle and was blocked, while dragging TOWARD it wrongly
//     admitted an ADD.
//
// writeTree's two nodes give one of each case for free: "1" is the edge's SOURCE
// (nm.outTargets/em.srcID == "1" — the branch that was broken) and "2" is the edge's
// TARGET (the branch that was always correct) — every assertion below runs against BOTH, so
// a regression on either side is caught.

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"testing"
)

// touchingBeadFor returns writeTree's single touching bead for nodeID ("1" or "2") and
// nodeID's own live centre, using the SAME dragTouchingBeads call production drives.
func touchingBeadFor(t *testing.T, md *MoveDispatch, nodeID string) (touchingBead, vec3) {
	t.Helper()
	nm, ok := md.mr.nodeMovers[nodeID]
	if !ok {
		t.Fatalf("no nodeMover for %s", nodeID)
	}
	prevPos := nodeWorldPos(nm.geom)
	beads := dragTouchingBeads(md, nm, prevPos)
	if len(beads) != 1 {
		t.Fatalf("%s: expected exactly one touching bead (one incident edge), got %d", nodeID, len(beads))
	}
	return beads[0], prevPos
}

// TestTouchingBeadSourceIsOneBeadLengthFromCentre asserts the touching bead's own SOURCE
// point sits ONE BEAD LENGTH (wire.BeadStepR) from the touching bead's own CENTRE — not
// nodeTorusOuterR(selfKind) (~5 bead lengths), which is what the broken isSource branch
// produced. writeTree's single edge puts "1" only ever on the source-side branch and
// "2" only ever on the target-side branch, so this exercises both.
func TestTouchingBeadSourceIsOneBeadLengthFromCentre(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)

	const eps = 1e-6
	for _, id := range []string{"1", "2"} {
		bead, _ := touchingBeadFor(t, md, id)
		got := bead.Centre.Sub(bead.Source).Length()
		if diff := got - wire.BeadStepR; diff > eps || diff < -eps {
			t.Fatalf("%s: touching bead source should be one bead length (wire.BeadStepR=%g) from its centre, got %g",
				id, wire.BeadStepR, got)
		}
		// selfTorusR is a genuinely different (larger) number in this fixture — pin that,
		// so a regression back to `prevPos + aimDir*selfTorusR` is caught even if
		// BeadStepR and selfTorusR ever happened to coincide for some future fixture.
		selfTorusR := nodeTorusOuterR(md.mr.nodeMovers[id].selfKind)
		if selfTorusR-wire.BeadStepR < 1.0 {
			t.Fatalf("%s: fixture's selfTorusR (%g) is too close to wire.BeadStepR (%g) to distinguish the two forms",
				id, selfTorusR, wire.BeadStepR)
		}
	}
}

// TestThirdAtRestIsOneBeadLengthNotSelfTorusR asserts |third| (bead_crud.go: nodeDestination
// - beadSource) measured with nodeDestination AT REST — the touching bead's own current
// centre, i.e. no drag at all — equals one bead length (wire.BeadStepR), not selfTorusR. The
// broken isSource branch made this ~selfTorusR (several bead lengths) on the source side, so
// |third| could never fall below the one-bead REMOVE threshold and REMOVE could never fire.
func TestThirdAtRestIsOneBeadLengthNotSelfTorusR(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)

	const eps = 1e-6
	for _, id := range []string{"1", "2"} {
		bead, _ := touchingBeadFor(t, md, id)
		third := bead.Centre.Sub(bead.Source)
		got := third.Length()
		if diff := got - wire.BeadStepR; diff > eps || diff < -eps {
			t.Fatalf("%s: |third| at rest should equal one bead length (wire.BeadStepR=%g), got %g (selfTorusR=%g)",
				id, wire.BeadStepR, got, nodeTorusOuterR(md.mr.nodeMovers[id].selfKind))
		}
	}
}

// TestAngleGateAdmitsAddAwayAndBlocksAddToward asserts beadCrudDecide's ADD angle gate reads
// the correct sense on BOTH the source-side and target-side touching bead: a drag whose
// nodeDestination is FURTHER from the touching bead's source than the bead's own centre
// (heading AWAY from the neighbour, opening a gap) must admit ADD; a drag whose
// nodeDestination heads back TOWARD the neighbour (past the bead, reducing the span below
// beadLen) must never admit ADD — only the broken isSource branch inverted this (beadVec
// pointed toward the neighbour instead of away), so dragging away was wrongly blocked and
// dragging toward was wrongly admitted.
func TestAngleGateAdmitsAddAwayAndBlocksAddToward(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)

	for _, id := range []string{"1", "2"} {
		bead, prevPos := touchingBeadFor(t, md, id)

		// dragVector = nodeDestination - prevPos, per beadCrudDecide's contract
		// (bead_crud.go). aimDir points from self TOWARD the neighbour, so moving AWAY
		// from the neighbour is -aimDir: extend outward well past the source in that
		// direction — |third| > beadLen and beadVec (beadCentre - beadSource, pointing
		// away from the neighbour when correct) forms an ACUTE angle with dragVector, so
		// ADD must be admitted.
		awayDest := prevPos.Sub(bead.AimDir.Scale(3 * wire.BeadStepR))
		awayDrag := awayDest.Sub(prevPos)
		verdict, _ := beadCrudDecide(bead.Source, bead.Centre, awayDest, awayDrag, wire.BeadStepR)
		if verdict != beadCrudAdd {
			t.Fatalf("%s: dragging AWAY from the neighbour should admit ADD, got verdict=%d", id, verdict)
		}

		// TOWARD: destination heads further toward the neighbour (+aimDir from prevPos),
		// far enough PAST the bead's own source point that |third| is > beadLen again on
		// the far side (so this isn't accidentally a REMOVE), but the angle between
		// beadVec and dragVector is still obtuse — ADD must be blocked (beadCrudNone).
		// sourceOffset is how far along aimDir the source sits from prevPos; overshooting
		// it by another 2 bead lengths guarantees |third| clears beadLen on the far side.
		sourceOffset := bead.Source.Sub(prevPos).Dot(bead.AimDir)
		towardDest := prevPos.Add(bead.AimDir.Scale(sourceOffset + 2*wire.BeadStepR))
		towardDrag := towardDest.Sub(prevPos)
		third := towardDest.Sub(bead.Source)
		if third.Length() <= wire.BeadStepR {
			t.Fatalf("%s: toward-destination fixture is degenerate — |third|=%g must exceed beadLen=%g to isolate the angle gate",
				id, third.Length(), wire.BeadStepR)
		}
		verdict, _ = beadCrudDecide(bead.Source, bead.Centre, towardDest, towardDrag, wire.BeadStepR)
		if verdict != beadCrudNone {
			t.Fatalf("%s: dragging TOWARD the neighbour (back across the bead) should block ADD, got verdict=%d", id, verdict)
		}
	}
}
