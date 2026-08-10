// Package beadcrud is the per-bead CRUD decision (PLAN.md "moving a node is CRUD on the
// edge beads touching it") plus the touching-bead resolution it judges, split out of
// nodes/Wiring (god-object decomposition): a would-be leaf that used to hold a
// *MoveDispatch/*nodeGeometry back-reference (dragTouchingBeads) now takes the values it
// reads — node id/kind, incident edge ids, each edge's neighbour id, partner centres, and
// neighbour kinds — as plain parameters, so this package needs nothing from nodes/Wiring
// and nodes/Wiring can import it with no cycle.
//
// bead_crud.go itself was always pure geometry: no node/edge state, no map lookups, no
// persistence, no side effects. Moving a node is decided independently, per touching bead,
// from that bead's own SOURCE point and the node's DESTINATION point — no solver, no
// enumeration across neighbours, no selection of one edge over another. The caller
// (nodes/Wiring's commit_node_move.go) supplies the real beadSource/beadCentre/
// nodeDestination for every touching bead and applies the verdicts; this file only judges
// ONE bead at a time.
package beadcrud

import (
	"math"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// vec3 is a local alias for wire.Vec3 — nodes/Wiring/vec_alias.go's own precedent for
// staying independent of any package that itself aliases the concrete type.
type vec3 = wire.Vec3

// BeadCrudVerdict is one touching bead's own answer to "does the node's drag change my
// edge's bead count".
type BeadCrudVerdict int

const (
	// BeadCrudNone: the touching bead's span to the node's destination still measures
	// one bead length — nothing changes.
	BeadCrudNone BeadCrudVerdict = iota
	// BeadCrudAdd: a gap has opened beyond this bead — a bead is added, and it becomes
	// the new touching bead.
	BeadCrudAdd
	// BeadCrudRemove: this bead no longer fits — it is removed, and the bead before it
	// becomes the touching bead.
	BeadCrudRemove
)

// BeadCrudDecide judges ONE touching bead.
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
//   - beadLen is one bead length (lattice.BeadStepR).
//
// third = nodeDestination - beadSource is the span the touching bead now has to occupy:
//
//   - |third| < beadLen  -> BeadCrudRemove — the bead no longer fits.
//   - |third| > beadLen  -> BeadCrudAdd, gated by the angle test below.
//   - |third| == beadLen -> BeadCrudNone — the drag changes nothing on this edge.
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
func BeadCrudDecide(beadSource, beadCentre, nodeDestination, dragVector vec3, beadLen float64) (BeadCrudVerdict, vec3) {
	third := nodeDestination.Sub(beadSource)
	tl := third.Length()
	switch {
	case tl < beadLen:
		return BeadCrudRemove, third
	case tl > beadLen:
		beadVec := beadCentre.Sub(beadSource)
		bl, dl := beadVec.Length(), dragVector.Length()
		if bl < 1e-12 || dl < 1e-12 {
			// Degenerate — no direction to gate against, so no add without a real
			// triangle to judge it by.
			return BeadCrudNone, third
		}
		cosA := beadVec.Dot(dragVector) / (bl * dl)
		if cosA < 0 {
			// angle > 90 degrees: drag is heading back across the bead.
			return BeadCrudNone, third
		}
		return BeadCrudAdd, third
	default:
		return BeadCrudNone, third
	}
}

// BeadCrudImpliedCentre computes the NODE's implied new centre from ONE touching bead's
// verdict — PLAN.md: "the node's position comes from the bead operation — NOT from the
// drag destination point... bead removed -> the node moves to take that bead's place;
// bead added -> the node moves away from the newly added bead's place." aimDir is the
// chain's own axis (the live unit direction from the node toward the neighbour, NEVER the
// drag direction); beadCentre is the touching bead's own CURRENT centre (before this
// event); beadLen is one bead length (lattice.BeadStepR). ok is false for BeadCrudNone — a
// "none" verdict implies no new position, only "unchanged".
//
//   - BeadCrudRemove: the node's new centre IS the removed bead's own centre — the node
//     takes that bead's place exactly.
//   - BeadCrudAdd: a new bead is inserted one bead length closer to the node than the old
//     touching bead, along aimDir (the "next chain position"); the node's new centre is
//     one bead length BEYOND that new bead, away from the neighbour (continuing along the
//     same axis, in the direction opposite aimDir).
func BeadCrudImpliedCentre(verdict BeadCrudVerdict, beadCentre, aimDir vec3, beadLen float64) (vec3, bool) {
	switch verdict {
	case BeadCrudRemove:
		return beadCentre, true
	case BeadCrudAdd:
		newBeadCentre := beadCentre.Sub(aimDir.Scale(beadLen))
		nodeCentre := newBeadCentre.Sub(aimDir.Scale(beadLen))
		return nodeCentre, true
	default:
		return vec3{}, false
	}
}

// BeadCrudDiag is ONE touching bead's full CRUD arithmetic, captured for observability
// only (task/log-node2-bead-crud breadcrumb) — commitNodeMoveLocal's "bead-crud"
// breadcrumb packs one of these per touching bead so a drag-time trace can show WHY a
// bead returned "none" instead of just that it did. Mirrors BeadCrudDecide's own
// arithmetic exactly (never a second, drifting copy of the verdict logic) but also keeps
// the intermediate values BeadCrudDecide discards: the angle-gate cosine and whether the
// gate is what blocked an otherwise-qualifying add, and the touching bead's own source
// distance from the node's previous position.
type BeadCrudDiag struct {
	NeighborID  string
	ThirdLen    float64 // |nodeDestination - beadSource|
	BeadLen     float64
	Verdict     BeadCrudVerdict
	CosAngle    float64 // angle-gate cosine; NaN when the gate was never evaluated (verdict != add-candidate-by-length)
	GateBlocked bool    // true iff |third|>beadLen but the angle gate turned it into "none"
	SourceDist  float64 // |beadSource - prevPos|
	Implied     vec3
	ImpliedOK   bool
}

// BeadCrudDiagnose judges ONE touching bead exactly like BeadCrudDecide, but also returns
// the intermediate angle-gate arithmetic for breadcrumb observability. DIAGNOSTIC ONLY —
// no caller depends on this for placement/movement; BeadCrudDecide remains the sole
// production judge.
func BeadCrudDiagnose(neighborID string, beadSource, beadCentre, aimDir, prevPos, nodeDestination, dragVector vec3, beadLen float64) BeadCrudDiag {
	third := nodeDestination.Sub(beadSource)
	tl := third.Length()
	d := BeadCrudDiag{
		NeighborID: neighborID,
		ThirdLen:   tl,
		BeadLen:    beadLen,
		CosAngle:   math.NaN(),
		SourceDist: beadSource.Sub(prevPos).Length(),
	}
	switch {
	case tl < beadLen:
		d.Verdict = BeadCrudRemove
	case tl > beadLen:
		beadVec := beadCentre.Sub(beadSource)
		bl, dl := beadVec.Length(), dragVector.Length()
		if bl < 1e-12 || dl < 1e-12 {
			d.Verdict = BeadCrudNone
		} else {
			cosA := beadVec.Dot(dragVector) / (bl * dl)
			d.CosAngle = cosA
			if cosA < 0 {
				d.Verdict = BeadCrudNone
				d.GateBlocked = true
			} else {
				d.Verdict = BeadCrudAdd
			}
		}
	default:
		d.Verdict = BeadCrudNone
	}
	if implied, ok := BeadCrudImpliedCentre(d.Verdict, beadCentre, aimDir, beadLen); ok {
		d.Implied, d.ImpliedOK = implied, true
	}
	return d
}

// BeadCrudResult is one touching bead's non-"none" verdict and the node centre it implies
// — ResolveBeadCrudMove's per-edge working set, kept for observability (breadcrumbs)
// alongside the resolved commit.
type BeadCrudResult struct {
	NeighborID string
	Verdict    BeadCrudVerdict
	Implied    vec3
}

// ResolveBeadCrudMove judges every touching bead against the SAME drag (nodeDestination,
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
//   - Exactly one touching bead signals a change: its BeadCrudImpliedCentre IS the node's
//     new committed centre.
//   - More than one touching bead signals a change: a node with several neighbours has
//     touching beads on several different chain axes, so their implied centres essentially
//     never coincide — that is the ORDINARY multi-neighbour case, not a conflict (an earlier
//     version treated disagreement as a conflict and held the node still, which made every
//     multi-neighbour node immovable — node 2 could not be dragged at all). The commit is
//     the implied centre with the SMALLEST displacement from prevPos among all verdicts —
//     a tie-break over the candidate centres the per-bead CRUD already produced, never an
//     average, and never the mouse target. Movement stays one bead at a time; an edge whose
//     verdict implied a larger step reaches it over successive pointer-move events instead
//     of in one jump (edgeStepCount re-counts against the live distance every commit).
func ResolveBeadCrudMove(beads []TouchingBead, prevPos, nodeDestination vec3, beadLen float64) (committed vec3, results []BeadCrudResult) {
	dragVector := nodeDestination.Sub(prevPos)
	for _, b := range beads {
		verdict, _ := BeadCrudDecide(b.Source, b.Centre, nodeDestination, dragVector, beadLen)
		if verdict == BeadCrudNone {
			continue
		}
		implied, ok := BeadCrudImpliedCentre(verdict, b.Centre, b.AimDir, beadLen)
		if !ok {
			continue
		}
		results = append(results, BeadCrudResult{NeighborID: b.NeighborID, Verdict: verdict, Implied: implied})
	}
	if len(results) == 0 {
		return prevPos, results
	}
	// Smallest-displacement is a tie-break, NOT a solver and NOT a selection of one edge's
	// axis to travel along: it reads only the candidate centres the per-bead CRUD already
	// produced, never the mouse target, never neighbour geometry, and it averages nothing.
	best := results[0]
	bestD := best.Implied.Sub(prevPos).Length()
	for _, r := range results[1:] {
		if d := r.Implied.Sub(prevPos).Length(); d < bestD {
			best, bestD = r, d
		}
	}
	return best.Implied, results
}
