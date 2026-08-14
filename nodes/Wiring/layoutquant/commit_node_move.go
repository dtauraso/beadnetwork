package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/topoderive"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func (lq *LayoutQuantizer) CommitNodeMoveLocal(nodeGeoms map[string]*nodeactor.NodeGeometry, edgeMovers map[string]*edgemover.EdgeMover, ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos spatial.Vec3, targetPolar *polar.Polar) {
	nodeID := nm.ID()
	edges := HeldEdges(edgeMovers)
	// A node knows its OWN polar and no one else's. It used to read its
	// neighbours' out of paths cached from their centre broadcasts; with
	// that gone there is nothing here to seed the map with.
	polars := map[string]polar.Polar{}

	// The sender's own triple wins when there is one. Cart2polar runs ONLY for
	// a position that came from the world in the first place — a pointer hit —
	// because that is the one case where there is no triple to lose. Deriving
	// it here regardless is what silently undid every composed constraint: the
	// fold answers in a canonical range and rewrites the other two components
	// to get there, so a phi pinned past the pole came back as a different
	// number standing at the same place (movemsg.Msg.TargetPolar).
	nodePolar := polar.Cart2polar(newPos.Sub(ui.SceneSphere.Center))
	if targetPolar != nil {
		nodePolar = *targetPolar
		newPos = ui.SceneSphere.Center.Add(polar.Polar2cart(nodePolar))
	}

	committedPos, committedPolar := lq.resolveCommittedPosition(ui, nm, newPos, nodePolar)

	polars[nodeID] = committedPolar
	reach := topoderive.ReachRFromPolar(polars, edges)

	// This node moved and the nodes at the other end of its edges did not, so
	// every side it touches loses the whole of Δ, component by component. It
	// applies that to its own sides here, and tells each partner the same Δ so
	// that partner can apply it to theirs — neither end reads the other's
	// position, and neither converts.
	delta := polar.Between(nm.ScenePolar(), committedPolar)
	nm.ShiftDeltasBy(delta)

	nm.ApplyCenter(committedPos, reach[nodeID])
	BroadcastToEdgesAndPartners(edgeMovers, nodeGeoms,
		map[string]spatial.Vec3{nodeID: committedPos},
		map[string]polar.Polar{nodeID: delta},
		nm.SendMove())

	nm.CommitQuantOffset(committedPolar)
}

// resolveCommittedPosition is where a move used to be adjusted before it
// committed. Bead-crud snapping needed the neighbour centres a node cached
// from their broadcasts, and the constraint trim needed the same thing, so
// with the cache gone a node commits the position it was given.
func (lq *LayoutQuantizer) resolveCommittedPosition(ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos spatial.Vec3, nodePolar polar.Polar) (committedPos spatial.Vec3, committedPolar polar.Polar) {
	return newPos, nodePolar
}
