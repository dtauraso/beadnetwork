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

func (lq *LayoutQuantizer) RootMove(ctx context.Context, nodeGeoms map[string]*nodeactor.NodeGeometry, nodeID string, target spatial.Vec3) bool {
	nm, ok := nodeGeoms[nodeID]
	if !ok {
		return false
	}

	// A drag moves THE NODE UNDER THE CURSOR. Dragging an input node does not
	// move its out-neighbours — not carried along, not re-solved around. They
	// stay where they are and the edges to them change, which is what dragging
	// that node is FOR.
	//
	// Two versions of moving them have now been wrong in the editor. Treating
	// them as standing still and re-imposing the shared length stretched the
	// shorter path and shoved its neighbour into other nodes; carrying them
	// rigidly moved nodes nobody was dragging. HeldOutNeighbors survives for the
	// LOAD-time hold only (build_move_dispatch.go, with a zero delta), which is
	// what corrects a layout that loads wrong.
	//
	// The drag arrives as a world point, because a pointer hits a plane in the
	// world. That is where it enters the polar system; from here down every
	// rule works on triples, and the node's own centre is the only one read.
	//
	// The theta of that triple is snapped to the nearest whole multiple of pi
	// the moment it is formed, before any rule reads it: a drag turns the node
	// half a turn about its pole or not at all, and nothing downstream ever
	// sees a partial turn to accumulate.
	delta := polar.SnapDeltaTheta(polar.Between(nm.ScenePolar(), polar.Cart2polar(target.Sub(nm.SceneCenter()))))
	delta = TrimDraggedNode(nm, delta)
	// Its neighbours are not moving, so keeping every outgoing path the same
	// length is a constraint on where THIS node may go.
	delta = TrimEqualOutLengths(nm, delta)

	// Dragging one of an input node's neighbours is the case that still moves
	// somebody else: the shared length is a constraint the dragged node cannot
	// satisfy alone, so its SIBLINGS take the length it just stated.
	for to, heldPoint := range HeldSiblings(nm, nodeGeoms, nodeID, delta) {
		other, ok := nodeGeoms[to]
		if !ok {
			continue
		}
		other.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: to,
			Target: other.SceneCenter().Add(polar.Polar2cart(heldPoint))})
	}

	nm.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: nodeID,
		Target: nm.SceneCenter().Add(polar.Polar2cart(polar.Compose(nm.ScenePolar(), delta)))})
	return true
}
