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
	delta := polar.Between(nm.ScenePolar(), polar.Cart2polar(target.Sub(nm.SceneCenter())))
	// An input node's own drag has its theta snapped to the nearest whole
	// multiple of pi the moment the triple is formed, before any rule reads it:
	// it turns half a turn about its own pole or not at all, and nothing
	// downstream ever sees a partial turn to accumulate. Every other node drags
	// with the theta the cursor asked for.
	//
	// It is the SAME gate the other rules on this node's own drag use — the
	// kind, not the id, so the rule belongs to what an input node is rather
	// than to which node happens to be first in this scene.
	if nm.SelfKind() == OutAngleKind {
		delta = polar.SnapDeltaTheta(delta)
	}
	delta = TrimDraggedNode(nm, delta)
	// Its neighbours are not moving, so keeping every outgoing path the same
	// length is a constraint on where THIS node may go.
	delta = TrimEqualOutLengths(nm, delta)

	// NOBODY ELSE MOVES. A drag sends exactly one message, to the node under
	// the cursor. Dragging one of an input node's out-targets used to also
	// move its siblings onto the length that drag had stated; the drag now
	// holds that length instead (TrimOutAngleDelta), so there is no length to
	// restate and no sibling to correct.
	nm.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: nodeID,
		Target: nm.SceneCenter().Add(polar.Polar2cart(polar.Compose(nm.ScenePolar(), delta)))})
	return true
}
