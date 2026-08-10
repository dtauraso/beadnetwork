// touching_beads.go — resolving, for a dragged node, the ONE bead on each incident edge
// that directly touches it.
//
// DragTouchingBeads used to take a *MoveDispatch and a *nodeGeometry directly (the
// back-reference this package's move fixes): it does not need a mover, it needs that
// mover's own topo/centres, so it now takes exactly the values it reads — node id/kind,
// incident edge ids, each edge's resolved neighbour id, this node's own partner-centre
// snapshot, and neighbour kinds. nodes/Wiring's touching_beads.go keeps a thin wrapper
// (dragTouchingBeads) that resolves those values from a live *MoveDispatch/*nodeGeometry
// and calls DragTouchingBeads — every existing call site in nodes/Wiring is unchanged.

package beadcrud

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// TouchingBead is the per-neighbour geometry commitNodeMoveLocal hands to BeadCrudDecide:
// the touching bead's own SOURCE point, its own CURRENT centre, and the chain's own AXIS
// (AimDir, the live unit direction from the dragged node toward the neighbour) — see
// DragTouchingBeads. AimDir is what the node's post-CRUD position is computed along
// (BeadCrudImpliedCentre); it is never the raw drag direction.
type TouchingBead struct {
	NeighborID string
	Source     vec3
	Centre     vec3
	AimDir     vec3
}

// DragTouchingBeads reads, for EVERY direct neighbour of the dragged node, the ONE bead
// on that edge that directly touches it — from the node's own live state only (its own
// id/kind, incident edge ids, each edge's resolved neighbour id, and its own
// partnerCenters/neighborKinds snapshots, the same live-copy maps chainBeads/
// requantizeLocalPolars already read). A neighbour with no live centre yet (never pushed
// an applyCenter) contributes no touching bead, same convention chainBeads uses for a
// target with no live partner center.
//
// neighborOf maps each incident edgeID to the OTHER endpoint's node id — resolved by the
// caller from its own edge-mover directory (nodes/Wiring's edgeMovers), since that
// directory is Wiring-internal and this package holds no back-reference to it. An edgeID
// missing from neighborOf (should not happen — every id in edgeIDs is expected to be a
// key) is skipped, same as a missing edgeMover was in the pre-move code.
//
// The touching bead's own centre sits at the SAME fixed offset from the dragged node
// regardless of which node owns the chain — tangency to its own torus falls out of the
// placement formula (nodes/Wiring's chain_beads.go), not from where its centre happens
// to be:
//
//	beadCentre = prevPos + aimDir*(selfTorusR + lattice.BeadTorusOuterR)
//
// where aimDir is the live unit direction from the dragged node toward the neighbour.
// What differs by ownership is the touching bead's SOURCE point (PLAN.md: "the previous
// bead's centre along its chain, or the chain origin on the neighbour's torus surface
// when it is the only bead") — NEVER the bead's own centre, which would be wrong by one
// bead:
//
//   - the dragged node is the edge's SOURCE: the touching bead is always chain index 0 —
//     it has no predecessor regardless of the current count, so its source is the
//     chain's own origin, the dragged node's own torus surface point:
//     beadSource = prevPos + aimDir*selfTorusR
//   - the dragged node is the edge's TARGET (an incoming edge): the touching bead is the
//     chain's LAST bead, owned and counted by the neighbour (edgeStepCount, same formula
//     chain_beads.go uses, mirrored here on the live distance). With more than one bead,
//     its predecessor is the bead one step back toward the neighbour:
//     beadSource = beadCentre + aimDir*lattice.BeadStepR
//     With exactly one bead, there is no predecessor bead — the chain's own origin is the
//     NEIGHBOUR's torus surface (neighborKinds gives the neighbour's kind, derived from
//     domain adjacency at load, so every direct neighbour has an entry):
//     beadSource = neighborCenter - aimDir*nodegeom.NodeTorusOuterR(neighborKind)
//
// dragTouchingBeads (nodes/Wiring) used to special-case which endpoint owns the edge to
// pick beadSource; it does not need to any more — neighborOf already resolved the
// neighbour id regardless of ownership, and the source-point rule above is the SAME
// whichever end owns the edge.
func DragTouchingBeads(nodeID, selfKind string, edgeIDs []string, neighborOf map[string]string, partnerCenters map[string]vec3, neighborKinds map[string]string, prevPos vec3) []TouchingBead {
	selfTorusR := nodegeom.NodeTorusOuterR(selfKind)
	out := make([]TouchingBead, 0, len(edgeIDs))
	for _, edgeID := range edgeIDs {
		neighborID, ok := neighborOf[edgeID]
		if !ok {
			continue
		}
		neighborCenter, ok := partnerCenters[neighborID]
		if !ok {
			continue
		}
		dist, aimDir, ok := nodegeom.EdgeCenterDistAndDir(prevPos, neighborCenter)
		if !ok {
			continue
		}
		beadCentre := prevPos.Add(aimDir.Scale(selfTorusR + lattice.BeadTorusOuterR))

		neighborKind := neighborKinds[neighborID]
		count := nodegeom.EdgeStepCount(dist, neighborKind, selfKind)
		var beadSource vec3
		if count >= 2 {
			beadSource = beadCentre.Add(aimDir.Scale(lattice.BeadStepR))
		} else {
			beadSource = neighborCenter.Sub(aimDir.Scale(nodegeom.NodeTorusOuterR(neighborKind)))
		}
		out = append(out, TouchingBead{NeighborID: neighborID, Source: beadSource, Centre: beadCentre, AimDir: aimDir})
	}
	return out
}
