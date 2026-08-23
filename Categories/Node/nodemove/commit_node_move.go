package nodemove

import (
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgetable"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (mv *NodeMover) CommitNodeMoveLocal(nodeGeoms map[string]*nodeactor.NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *nodeactor.NodeGeometry, committedIdx polarindex.Index) {
	nodeID := nm.ID()

	deltaIdx := polarindex.Delta(committedIdx, nm.ComposedIndex())
	nm.Deltas().ShiftSelfBy(deltaIdx)

	nm.ApplyCenter(committedIdx)
	committedPos := nm.SceneCenter().Add(nodeactor.Vec3(polar.Polar2cart(polarindex.ToPolar(committedIdx, nm.Constants()))))
	BroadcastToPartners(edgeTable, nodeGeoms,
		map[string]Vec3{nodeID: Vec3(committedPos)},
		map[string]polarindex.Offset{nodeID: deltaIdx},
		nm.SendMove())

	nm.CommitIndex()
}
