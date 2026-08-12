package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func BroadcastToEdgesAndPartners(edgeMovers map[string]*edgemover.EdgeMover, nodeGeoms map[string]*nodeactor.NodeGeometry, newCenters map[string]spatial.Vec3, enqueue func(id string, msg movemsg.Msg)) {

	for edgeID, em := range edgeMovers {
		eps := map[string]spatial.Vec3{}
		if c, ok := newCenters[em.SrcID()]; ok {
			eps[em.SrcID()] = c
		}
		if c, ok := newCenters[em.DstID()]; ok {
			eps[em.DstID()] = c
		}
		if len(eps) == 0 {
			continue
		}
		enqueue(edgeID, movemsg.Msg{Kind: movemsg.KindCenters, Centers: eps})
	}

	partners := map[string]string{}
	for _, em := range edgeMovers {
		if _, moved := newCenters[em.SrcID()]; moved {
			if _, alsoMoved := newCenters[em.DstID()]; !alsoMoved {
				partners[em.DstID()] = em.SrcID()
			}
		}
		if _, moved := newCenters[em.DstID()]; moved {
			if _, alsoMoved := newCenters[em.SrcID()]; !alsoMoved {
				partners[em.SrcID()] = em.DstID()
			}
		}
	}
	for partnerID, movedID := range partners {
		if _, ok := nodeGeoms[partnerID]; !ok {
			continue
		}

		enqueue(partnerID, movemsg.Msg{Kind: movemsg.KindCenter, NodeID: partnerID, Center: nil,
			SenderID: movedID})
	}
}
