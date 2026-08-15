package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgetable"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func (lq *LayoutQuantizer) CommitNodeMoveLocal(nodeGeoms map[string]*nodeactor.NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *nodeactor.NodeGeometry, committedIdx polarindex.Index) {
	nodeID := nm.ID()

	deltaIdx := polarindex.Delta(committedIdx, nm.ComposedIndex())
	nm.ShiftDeltasBy(deltaIdx)

	nm.ApplyCenter(committedIdx)
	committedPos := nm.SceneCenter().Add(polar.Polar2cart(polarindex.ToPolar(committedIdx, nm.Constants())))
	BroadcastToPartners(edgeTable, nodeGeoms,
		map[string]spatial.Vec3{nodeID: committedPos},
		map[string]polarindex.Offset{nodeID: deltaIdx},
		nm.SendMove())

	nm.CommitIndex()
}
