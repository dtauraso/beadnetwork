package Wiring

// bead_crud.go — the per-bead CRUD decision (PLAN.md "moving a node is CRUD on the edge
// beads touching it"). Moving a node is decided independently, per touching bead, from
// that bead's own SOURCE point and the node's DESTINATION point — no solver, no
// enumeration across neighbours, no selection of one edge over another. This file is
// pure geometry: no node/edge state, no map lookups, no persistence, no side effects.
// The caller (quantized_move.go's commitNodeMoveLocal) supplies the real beadSource/
// beadCentre/nodeDestination for every touching bead and applies the verdicts; this file
// only judges ONE bead at a time.

// beadCrudVerdict is one touching bead's own answer to "does the node's drag change my
// edge's bead count".
type beadCrudVerdict int

const (
	// beadCrudNone: the touching bead's span to the node's destination still measures
	// one bead length — nothing changes.
	beadCrudNone beadCrudVerdict = iota
	// beadCrudAdd: a gap has opened beyond this bead — a bead is added, and it becomes
	// the new touching bead.
	beadCrudAdd
	// beadCrudRemove: this bead no longer fits — it is removed, and the bead before it
	// becomes the touching bead.
	beadCrudRemove
)

// beadCrudDecide judges ONE touching bead.
//
//   - beadSource is that bead's own SOURCE point: the previous bead's centre along its
//     chain, or the chain origin on the neighbour's torus surface when it is the only
//     bead. NEVER the touching bead's own centre — using the bead's own centre instead
//     is wrong by one bead (PLAN.md).
//   - beadCentre is the touching bead's own CURRENT centre.
//   - nodeDestination is the dragged node's destination point — the SAME point for every
//     touching bead judged this event.
//   - dragVector is the node's OWN polar vector v ("the two drag points give the node's
//     polar vector v", PLAN.md) — the node's own previous position to its destination.
//     ONE vector, computed once by the caller and passed unchanged to every touching
//     bead; it is never derived per-bead.
//   - beadLen is one bead length (wire.BeadStepR).
//
// third = nodeDestination - beadSource is the span the touching bead now has to occupy:
//
//   - |third| < beadLen  -> beadCrudRemove — the bead no longer fits.
//   - |third| > beadLen  -> beadCrudAdd, gated by the angle test below.
//   - |third| == beadLen -> beadCrudNone — the drag changes nothing on this edge.
//
// The angle gate applies to ADD only, never to REMOVE: the angle between dragVector (v)
// and the edge-bead vector (beadSource -> beadCentre) —
//
//   - > 90 degrees  -> add nothing. The node did not move far enough AWAY from the bead
//     to open a gap beyond it; an obtuse angle means the drag is heading back across the
//     bead instead.
//   - <= 90 degrees -> the add may proceed (the |third| test above already put it here).
//
// third is returned alongside the verdict purely for the caller's own observability
// (breadcrumbs); this function makes no other use of it and has no side effects.
func beadCrudDecide(beadSource, beadCentre, nodeDestination, dragVector vec3, beadLen float64) (beadCrudVerdict, vec3) {
	third := nodeDestination.Sub(beadSource)
	tl := third.Length()
	switch {
	case tl < beadLen:
		return beadCrudRemove, third
	case tl > beadLen:
		beadVec := beadCentre.Sub(beadSource)
		bl, dl := beadVec.Length(), dragVector.Length()
		if bl < 1e-12 || dl < 1e-12 {
			// Degenerate — no direction to gate against, so no add without a real
			// triangle to judge it by.
			return beadCrudNone, third
		}
		cosA := beadVec.Dot(dragVector) / (bl * dl)
		if cosA < 0 {
			// angle > 90 degrees: drag is heading back across the bead.
			return beadCrudNone, third
		}
		return beadCrudAdd, third
	default:
		return beadCrudNone, third
	}
}

// beadCrudImpliedCentre computes the NODE's implied new centre from ONE touching bead's
// verdict — PLAN.md: "the node's position comes from the bead operation — NOT from the
// drag destination point... bead removed -> the node moves to take that bead's place;
// bead added -> the node moves away from the newly added bead's place." aimDir is the
// chain's own axis (the live unit direction from the node toward the neighbour, NEVER the
// drag direction); beadCentre is the touching bead's own CURRENT centre (before this
// event); beadLen is one bead length (wire.BeadStepR). ok is false for beadCrudNone — a
// "none" verdict implies no new position, only "unchanged".
//
//   - beadCrudRemove: the node's new centre IS the removed bead's own centre — the node
//     takes that bead's place exactly.
//   - beadCrudAdd: a new bead is inserted one bead length closer to the node than the old
//     touching bead, along aimDir (the "next chain position"); the node's new centre is
//     one bead length BEYOND that new bead, away from the neighbour (continuing along the
//     same axis, in the direction opposite aimDir).
func beadCrudImpliedCentre(verdict beadCrudVerdict, beadCentre, aimDir vec3, beadLen float64) (vec3, bool) {
	switch verdict {
	case beadCrudRemove:
		return beadCentre, true
	case beadCrudAdd:
		newBeadCentre := beadCentre.Sub(aimDir.Scale(beadLen))
		nodeCentre := newBeadCentre.Sub(aimDir.Scale(beadLen))
		return nodeCentre, true
	default:
		return vec3{}, false
	}
}

// beadCrudResult is one touching bead's non-"none" verdict and the node centre it implies
// — resolveBeadCrudMove's per-edge working set, kept for observability (breadcrumbs,
// conflict reporting) alongside the resolved commit.
type beadCrudResult struct {
	NeighborID string
	Verdict    beadCrudVerdict
	Implied    vec3
}

// resolveBeadCrudMove judges every touching bead against the SAME drag (nodeDestination,
// dragVector = nodeDestination-prevPos) and resolves the node's single committed centre —
// the ONE place this package decides how multiple touching beads' verdicts become one
// Cartesian point, so commitNodeMoveLocal (production) and quantizedDragTarget (the test
// oracle, subtree_persist_test.go) can share the exact same resolution instead of two
// copies that could drift.
//
//   - No touching bead signals a change (every verdict "none", or there are no touching
//     beads at all): the node does not move — prevPos, unchanged. (A node with NO incident
//     edges is a different case, handled by the caller before this is reached: it follows
//     the raw target directly, matching the historic free-node behaviour.)
//   - Exactly one touching bead signals a change: its beadCrudImpliedCentre IS the node's
//     new committed centre.
//   - More than one touching bead signals a change: if every one of their implied centres
//     agrees (the same point, within float tolerance), that shared point is the commit —
//     this is the ordinary multi-edge case (PLAN.md's dK=(-1,+1,0) example, one edge adding
//     while another removes, both driven by the SAME node move). If they DISAGREE — no two
//     touching beads can imply two different single points for the one node centre — this
//     is a genuine conflict in the per-bead verdicts: NOT resolved by averaging, nearest-
//     to-cursor, or falling back to the drag target (PLAN.md forbids all three). The node
//     holds its position and conflict=true, so the caller can make the conflict observable
//     (breadcrumb) rather than silently picking a point.
func resolveBeadCrudMove(beads []touchingBead, prevPos, nodeDestination vec3, beadLen float64) (committed vec3, results []beadCrudResult, conflict bool) {
	dragVector := nodeDestination.Sub(prevPos)
	for _, b := range beads {
		verdict, _ := beadCrudDecide(b.Source, b.Centre, nodeDestination, dragVector, beadLen)
		if verdict == beadCrudNone {
			continue
		}
		implied, ok := beadCrudImpliedCentre(verdict, b.Centre, b.AimDir, beadLen)
		if !ok {
			continue
		}
		results = append(results, beadCrudResult{NeighborID: b.NeighborID, Verdict: verdict, Implied: implied})
	}
	switch len(results) {
	case 0:
		return prevPos, results, false
	case 1:
		return results[0].Implied, results, false
	default:
		for _, r := range results[1:] {
			if r.Implied.Sub(results[0].Implied).Length() > 1e-6 {
				return prevPos, results, true
			}
		}
		return results[0].Implied, results, false
	}
}
