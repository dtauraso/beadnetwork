package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/topoderive"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func (lq *LayoutQuantizer) CommitNodeMoveLocal(nodeGeoms map[string]*nodeactor.NodeGeometry, edgeMovers map[string]*edgemover.EdgeMover, ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos spatial.Vec3) {
	nodeID := nm.ID()
	edges := HeldEdges(edgeMovers)
	// A node knows its OWN polar and no one else's. It used to read its
	// neighbours' out of paths cached from their centre broadcasts; with
	// that gone there is nothing here to seed the map with.
	polars := map[string]polar.Polar{}

	nodePolar := polar.Cart2polar(newPos.Sub(ui.SceneSphere.Center))

	committedPos, committedPolar := lq.resolveCommittedPosition(ui, nm, newPos, nodePolar)

	polars[nodeID] = committedPolar
	reach := topoderive.ReachRFromPolar(polars, edges)

	nm.ApplyCenter(committedPos, reach[nodeID])
	BroadcastToEdgesAndPartners(edgeMovers, nodeGeoms, map[string]spatial.Vec3{nodeID: committedPos}, nm.SendMove())

	nm.CommitQuantOffset(committedPolar)
}

// resolveCommittedPosition is where a move used to be adjusted before it
// committed. Bead-crud snapping needed the neighbour centres a node cached
// from their broadcasts, and the constraint trim needed the same thing, so
// with the cache gone a node commits the position it was given.
func (lq *LayoutQuantizer) resolveCommittedPosition(ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos spatial.Vec3, nodePolar polar.Polar) (committedPos spatial.Vec3, committedPolar polar.Polar) {
	return newPos, nodePolar
}
