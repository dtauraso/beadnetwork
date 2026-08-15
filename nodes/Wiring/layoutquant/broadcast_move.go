package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgetable"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func BroadcastToPartners(edges map[string]*edgetable.Edge, nodeGeoms map[string]*nodeactor.NodeGeometry, newCenters map[string]spatial.Vec3, moveDeltas map[string]polarindex.Index, enqueue func(id string, msg movemsg.Msg)) {

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

		msg := movemsg.Msg{Kind: movemsg.KindCenter, NodeID: partnerID, Center: nil, SenderID: movedID}
		if d, ok := moveDeltas[movedID]; ok {
			msg.Delta = &d
		}
		enqueue(partnerID, msg)
	}
}
