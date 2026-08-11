// touching_beads.go — the live-mover adapter for beadcrud.DragTouchingBeads (god-object
// decomposition: dragTouchingBeads used to hold a *MoveDispatch/*nodeGeometry
// back-reference itself; the pure resolution now lives in nodes/Wiring/beadcrud and takes
// exactly the values it reads, so this file's only job is resolving those values from a
// live mover and calling it).

package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/beadcrud"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
)

// dragTouchingBeads resolves nm's own incident-edge neighbours (via mr.edgeMovers,
// Wiring-internal — beadcrud holds no back-reference to it) into an edgeID->neighborID
// map, then calls beadcrud.DragTouchingBeads with nm's own id/kind/edgeIDs/
// partnerCenters/neighborKinds — see that function's doc comment for the full placement
// rule. Same values, same result, as when this function did the resolution itself.
func dragTouchingBeads(mr *moverRegistry, nm *nodeactor.NodeGeometry, prevPos vec3) []beadcrud.TouchingBead {
	edgeIDs := nm.EdgeIDs()
	selfID := nm.ID()
	neighborOf := make(map[string]string, len(edgeIDs))
	for _, edgeID := range edgeIDs {
		em, ok := mr.edgeMovers[edgeID]
		if !ok {
			continue
		}
		neighborID := em.SrcID()
		if neighborID == selfID {
			neighborID = em.DstID()
		}
		neighborOf[edgeID] = neighborID
	}
	return beadcrud.DragTouchingBeads(selfID, nm.SelfKind(), edgeIDs, neighborOf, nm.PartnerCenters(), nm.NeighborKinds(), prevPos)
}
