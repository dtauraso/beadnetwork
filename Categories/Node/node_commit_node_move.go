package Node

import (
	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgetable"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (mv *NodeMover) CommitNodeMoveLocal(nodeGeoms map[string]*NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *NodeGeometry, committedIdx polarindex.Index) {
	nodeID := nm.ID()

	deltaIdx := polarindex.Delta(committedIdx, nm.ComposedIndex())
	nm.Deltas().ShiftSelfBy(deltaIdx)

	nm.ApplyCenter(committedIdx)
	committedPos := nodegeom.WorldPosAt(nm.SceneCenter(), committedIdx, nm.Constants())
	BroadcastToPartners(edgeTable, nodeGeoms,
		map[string]nodegeom.Vec3{nodeID: nodegeom.Vec3(committedPos)},
		map[string]polarindex.Offset{nodeID: deltaIdx},
		nm.Msg().SendMove())

	nm.CommitIndex()
}
