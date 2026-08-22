package nodemove

import (
	"github.com/dtauraso/wirefold/src/Node/Edge/edgetable"
	"github.com/dtauraso/wirefold/src/Node/nodeactor"
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
	"github.com/dtauraso/wirefold/src/spatial"
)

func (mv *NodeMover) CommitNodeMoveLocal(nodeGeoms map[string]*nodeactor.NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *nodeactor.NodeGeometry, committedIdx polarindex.Index) {
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
