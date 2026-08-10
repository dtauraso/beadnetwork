// touching_beads.go — resolving, for a dragged node, the ONE bead on each incident edge
// that directly touches it.
//
// Split out of quantized_move.go (god-object decomposition, pure move — no logic
// changes): this is the per-neighbour geometry commitNodeMoveLocal hands to
// beadCrudDecide, kept apart from the commit path itself and from held-state/broadcast.

package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// touchingBead is the per-neighbour geometry commitNodeMoveLocal hands to beadCrudDecide:
// the touching bead's own SOURCE point, its own CURRENT centre, and the chain's own AXIS
// (AimDir, the live unit direction from nm toward the neighbour) — see dragTouchingBeads.
// AimDir is what the node's post-CRUD position is computed along (beadCrudImpliedCentre);
// it is never the raw drag direction.
type touchingBead struct {
	NeighborID string
	Source     vec3
	Centre     vec3
	AimDir     vec3
}

// dragTouchingBeads reads, for EVERY direct neighbour of the dragged node nm, the ONE bead
// on that edge that directly touches nm — from nm's own live state only (its own kind,
// prevPos, and partnerCenters, the same live-copy map chainBeads/requantizeLocalPolars
// already read). A neighbour with no live centre yet (never pushed an applyCenter)
// contributes no touching bead, same convention chainBeads uses for a target with no live
// partner center.
//
// The touching bead's own centre sits at the SAME fixed offset from nm regardless of which
// node owns the chain — tangency to nm's own torus falls out of the placement formula
// (chain_beads.go), not from where nm's centre happens to be:
//
//	beadCentre = prevPos + aimDir*(selfTorusR + lattice.BeadTorusOuterR)
//
// where aimDir is the live unit direction from nm toward the neighbour. What differs by
// ownership is the touching bead's SOURCE point (PLAN.md: "the previous bead's centre
// along its chain, or the chain origin on the neighbour's torus surface when it is the
// only bead") — NEVER the bead's own centre, which would be wrong by one bead:
//
//   - nm is the edge's SOURCE (nm.outTargets contains the neighbour): the touching bead is
//     always chain index 0 — it has no predecessor regardless of the current count, so its
//     source is the chain's own origin, nm's own torus surface point:
//     beadSource = prevPos + aimDir*selfTorusR
//   - nm is the edge's TARGET (an incoming edge): the touching bead is the chain's LAST
//     bead, owned and counted by the neighbour (edgeStepCount, same formula chain_beads.go
//     uses, mirrored here on the live distance). With more than one bead, its predecessor
//     is the bead one step back toward the neighbour:
//     beadSource = beadCentre + aimDir*lattice.BeadStepR
//     With exactly one bead, there is no predecessor bead — the chain's own origin is the
//     NEIGHBOUR's torus surface (nm.topo.neighborKinds gives the neighbour's kind, derived from
//     domain adjacency at load — see build.go — so every direct neighbour has an entry):
//     beadSource = neighborCenter - aimDir*nodegeom.NodeTorusOuterR(neighborKind)
func dragTouchingBeads(md *MoveDispatch, nm *nodeGeometry, prevPos vec3) []touchingBead {
	nodeID := nm.id
	selfTorusR := nodegeom.NodeTorusOuterR(nm.selfKind)
	out := make([]touchingBead, 0, len(nm.topo.edgeIDs))
	for _, edgeID := range nm.topo.edgeIDs {
		em, ok := md.mr.edgeMovers[edgeID]
		if !ok {
			continue
		}
		neighborID := em.srcID
		isSource := em.srcID == nodeID
		if isSource {
			neighborID = em.dstID
		}
		neighborCenter, ok := nm.topo.partnerCenters[neighborID]
		if !ok {
			continue
		}
		dist, aimDir, ok := nodegeom.EdgeCenterDistAndDir(prevPos, neighborCenter)
		if !ok {
			continue
		}
		beadCentre := prevPos.Add(aimDir.Scale(selfTorusR + lattice.BeadTorusOuterR))

		// A touching bead's SOURCE POINT is the previous bead's centre along its chain, or
		// the chain origin on the NEIGHBOUR's torus surface when it is the only bead
		// (MODEL.md's bead-CRUD section) — the SAME rule whichever end of the edge this
		// node happens to own. Which endpoint stores the edge changes nothing about where
		// the beads sit.
		//
		// This used to special-case `isSource` to `prevPos + aimDir*selfTorusR` — a point
		// on THIS node's own torus surface, which is neither of the two the model allows.
		// It broke the rule twice over:
		//   - |third| at rest became selfTorusR (~5 bead lengths, ~44.8 world units) rather
		//     than one bead length, so |third| never fell below the one-bead threshold and
		//     REMOVE could never fire at all; and
		//   - beadVec came out as +aimDir*BeadTorusOuterR, pointing TOWARD the neighbour
		//     instead of away from it, inverting the angle gate — dragging AWAY (which
		//     should open a gap and add a bead) scored > 90 degrees and was blocked, while
		//     dragging toward admitted an add whose implied centre sat ~31 world units
		//     away. Hence: drag far in most directions and nothing happens; drag the other
		//     way and the node jumps.
		neighborKind := nm.topo.neighborKinds[neighborID]
		count := nodegeom.EdgeStepCount(dist, neighborKind, nm.selfKind)
		var beadSource vec3
		if count >= 2 {
			beadSource = beadCentre.Add(aimDir.Scale(lattice.BeadStepR))
		} else {
			beadSource = neighborCenter.Sub(aimDir.Scale(nodegeom.NodeTorusOuterR(neighborKind)))
		}
		out = append(out, touchingBead{NeighborID: neighborID, Source: beadSource, Centre: beadCentre, AimDir: aimDir})
	}
	return out
}
