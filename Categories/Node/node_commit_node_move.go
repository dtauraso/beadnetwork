package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgetable"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (mv *NodeMover) CommitNodeMoveLocal(nodeGeoms map[string]*NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *NodeGeometry, committedIdx polarindex.Index) {
	mv.applyNodeMoveLocal(nodeGeoms, edgeTable, nm, committedIdx)
	nm.CommitIndex()
}

func (mv *NodeMover) ApplyDerivedNodeMove(nodeGeoms map[string]*NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *NodeGeometry, committedIdx polarindex.Index) {
	nm.OutEdges().SuppressDeltaPersist()
	mv.applyNodeMoveLocal(nodeGeoms, edgeTable, nm, committedIdx)
}

func (mv *NodeMover) applyNodeMoveLocal(nodeGeoms map[string]*NodeGeometry, edgeTable map[string]*edgetable.Edge, nm *NodeGeometry, committedIdx polarindex.Index) {
	nodeID := nm.ID()

	deltaIdx := polarindex.Delta(committedIdx, nm.ComposedIndex())
	nm.Deltas().ShiftSelfBy(deltaIdx)

	nm.ApplyCenter(committedIdx)
	committedPos := WorldPosAt(nm.SceneCenter(), committedIdx, nm.Constants())
	BroadcastToPartners(edgeTable, nodeGeoms,
		map[string]Vec3{nodeID: Vec3(committedPos)},
		map[string]polarindex.Offset{nodeID: deltaIdx},
		nm.Msg().SendMove())
}
