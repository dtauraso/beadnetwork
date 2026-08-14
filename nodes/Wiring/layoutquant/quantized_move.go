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
	// world. That is where it enters the polar system, and it is ALL that
	// happens here: the hit becomes Δ against the dragged node's own centre,
	// and the node is sent that triple untrimmed.
	//
	// No rule is read here and none may be added. The trims used to run in this
	// function, on the pointer's goroutine, reaching into the node's own kind,
	// orbit rule and edge sides to decide them before the node had heard
	// anything. They are now the node's own (nodeactor.NodeGeometry.TrimOwnDrag),
	// which is a statement this seam can enforce rather than describe: this
	// package imports nodeactor, so a rule written here could reach the node's
	// state, and one written there cannot reach anyone else's.
	delta := polar.Between(nm.ScenePolar(), polar.Cart2polar(target.Sub(nm.SceneCenter())))

	// NOBODY ELSE MOVES. A drag sends exactly one message, to the node under
	// the cursor. Dragging one of an input node's out-targets used to also
	// move its siblings onto the length that drag had stated; the drag now
	// holds that length instead (its own polar.OrbitRule), so there is no length to
	// restate and no sibling to correct.
	//
	// Δ travels as the three numbers, not as a point. A world position would
	// hand the node a place and make it guess the triple back, and the guess is
	// canonical where the composition need not be (movemsg.Msg.Delta).
	nm.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: nodeID, Delta: &delta})
	return true
}
