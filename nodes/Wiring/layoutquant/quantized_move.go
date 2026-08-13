package layoutquant

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

type LayoutQuantizer struct {
	QuantizedLayout bool
}

func HeldCenters(nodeGeoms map[string]*nodeactor.NodeGeometry, centerOf func(id string) (spatial.Vec3, bool)) map[string]spatial.Vec3 {
	out := make(map[string]spatial.Vec3, len(nodeGeoms))
	for id := range nodeGeoms {
		if c, ok := centerOf(id); ok {
			out[id] = c
		}
	}
	return out
}

func HeldEdges(edgeMovers map[string]*edgemover.EdgeMover) []polar.SphereEdge {
	edges := make([]polar.SphereEdge, 0, len(edgeMovers))
	for _, em := range edgeMovers {
		edges = append(edges, polar.SphereEdge{Source: em.SrcID(), Target: em.DstID()})
	}
	return edges
}

func (lq *LayoutQuantizer) RootMove(ctx context.Context, nodeGeoms map[string]*nodeactor.NodeGeometry, centerOf func(string) (spatial.Vec3, bool), nodeID string, target spatial.Vec3) bool {
	nm, ok := nodeGeoms[nodeID]
	if !ok {
		return false
	}

	// A drag decides everything that moves. The dragged node's own
	// constraints trim where IT may go, and if it is a node that holds
	// constraints over others, the same drag carries them — a neighbour
	// never moves itself to satisfy someone else's rule.
	target = TrimDraggedNode(nm, centerOf, nm.WorldCenter(), target)

	held := HeldOutNeighbors(nm, centerOf, target)
	for to, pos := range HeldSiblings(nm, nodeGeoms, centerOf, nodeID, target) {
		if held == nil {
			held = map[string]spatial.Vec3{}
		}
		held[to] = pos
	}
	for to, heldPos := range held {
		other, ok := nodeGeoms[to]
		if !ok {
			continue
		}
		other.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: to, Target: heldPos})
	}

	nm.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: nodeID, Target: target})
	return true
}
