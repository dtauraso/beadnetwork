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
