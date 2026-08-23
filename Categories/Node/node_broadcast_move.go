package Node

import (
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgetable"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func BroadcastToPartners(edges map[string]*edgetable.Edge, nodeGeoms map[string]*NodeGeometry, newCenters map[string]nodegeom.Vec3, moveDeltas map[string]polarindex.Offset, enqueue func(id string, msg Msg)) {

	partners := map[string]string{}
	for _, e := range edges {
		if _, moved := newCenters[e.SrcID()]; moved {
			if _, alsoMoved := newCenters[e.DstID()]; !alsoMoved {
				partners[e.DstID()] = e.SrcID()
			}
		}
		if _, moved := newCenters[e.DstID()]; moved {
			if _, alsoMoved := newCenters[e.SrcID()]; !alsoMoved {
				partners[e.SrcID()] = e.DstID()
			}
		}
	}
	for partnerID, movedID := range partners {
		if _, ok := nodeGeoms[partnerID]; !ok {
			continue
		}

		moved := NeighborMoved{SenderID: movedID}
		if d, ok := moveDeltas[movedID]; ok {
			moved.Delta = &d
		}
		enqueue(partnerID, Msg{NodeID: partnerID, Body: moved})
	}
}
