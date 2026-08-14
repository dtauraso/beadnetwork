package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgetable"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/topoderive"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func (lq *LayoutQuantizer) CommitNodeMoveLocal(nodeGeoms map[string]*nodeactor.NodeGeometry, edgeTable map[string]*edgetable.Edge, ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos spatial.Vec3, targetPolar *polar.Polar) {
	nodeID := nm.ID()
	edges := HeldEdges(edgeTable)
	polars := map[string]polar.Polar{}

	nodePolar := polar.Cart2polar(newPos.Sub(ui.SceneSphere.Center))
	if targetPolar != nil {
		nodePolar = *targetPolar
		newPos = ui.SceneSphere.Center.Add(polar.Polar2cart(nodePolar))
	}

	committedPos, committedPolar := lq.resolveCommittedPosition(ui, nm, newPos, nodePolar)

	polars[nodeID] = committedPolar
	reach := topoderive.ReachRFromPolar(polars, edges)

	delta := polar.Between(nm.ScenePolar(), committedPolar)
	nm.ShiftDeltasBy(delta)

	nm.ApplyCenter(committedPos, reach[nodeID])
	BroadcastToPartners(edgeTable, nodeGeoms,
		map[string]spatial.Vec3{nodeID: committedPos},
		map[string]polar.Polar{nodeID: delta},
		nm.SendMove())

	nm.CommitQuantOffset(committedPolar)
}

func (lq *LayoutQuantizer) resolveCommittedPosition(ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos spatial.Vec3, nodePolar polar.Polar) (committedPos spatial.Vec3, committedPolar polar.Polar) {
	return newPos, nodePolar
}
