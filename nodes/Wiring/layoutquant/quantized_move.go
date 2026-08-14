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

	// The theta an input node's drag has ASKED FOR and not yet been given.
	//
	// A drag's theta may only be a whole multiple of pi, and RootMove runs once
	// per POINTER-MOVE EVENT rather than once per drag, so a single event's
	// theta is far too small to be one of those multiples and rounds to zero.
	// Rounding it away each event is what made the node sit still through the
	// arc and then jump: every step was DISCARDED, so the only turns that ever
	// happened were the rare single events big enough to cross half a turn on
	// their own.
	//
	// Keeping the remainder is what turns that back into a drag. The node still
	// moves only in whole multiples of pi — the rule is untouched — but it takes
	// one as soon as the cursor has asked for half a turn IN TOTAL.
	//
	// This is the gesture goroutine's own state: RootMove is called from there
	// and nowhere else (gesture_handlers.go, gesture_graph.go), so the node's
	// own geometry is not written by a stranger to hold it.
	unturnedTheta float64
	// Which node the remainder was gathered for, so a drag on another node
	// starts from zero instead of inheriting it.
	unturnedNode string
}

// turnAsked answers with the theta this event's drag may actually take: the
// whole multiple of pi nearest to everything asked for and not yet given,
// with whatever is left over held for the next event.
//
// asked is one pointer event's worth of theta, which is why it is added rather
// than rounded on its own — see unturnedTheta.
func (lq *LayoutQuantizer) turnAsked(nodeID string, asked float64) float64 {
	if lq.unturnedNode != nodeID {
		lq.unturnedNode, lq.unturnedTheta = nodeID, 0
	}
	lq.unturnedTheta += asked
	turn := polar.NearestHalfTurn(lq.unturnedTheta)
	lq.unturnedTheta -= turn
	return turn
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
	// it turns half a turn about its own pole or not at all. What an event asked
	// for and did not get is KEPT rather than dropped, so the turn happens once
	// the cursor has asked for it in total. Every other node drags with the
	// theta the cursor asked for.
	//
	// It is the SAME gate the other rules on this node's own drag use — the
	// kind, not the id, so the rule belongs to what an input node is rather
	// than to which node happens to be first in this scene.
	if nm.SelfKind() == OutAngleKind {
		delta.Theta = lq.turnAsked(nodeID, delta.Theta)
	}
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
