package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/beadcrud"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func DragTouchingBeads(edgeMovers map[string]*edgemover.EdgeMover, nm *nodeactor.NodeGeometry, prevPos spatial.Vec3) []beadcrud.TouchingBead {
	edgeIDs := nm.EdgeIDs()
	selfID := nm.ID()
	neighborOf := make(map[string]string, len(edgeIDs))
	for _, edgeID := range edgeIDs {
		em, ok := edgeMovers[edgeID]
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
