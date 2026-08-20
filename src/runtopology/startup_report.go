package runtopology

import (
	W "github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

func emitStartupBreadcrumbs(md *W.MoveDispatch, scenePath string, nodeCount int) {

	md.UI.EmitBreadcrumb(B.RowEvent{
		Label: B.BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(nodeCount), Text: scenePath,
	})
}

func checkRowSeedCount(md *W.MoveDispatch, nodeCount int) {
	if len(md.GS.NodeSeedsFn()) != nodeCount {

		md.UI.EmitBreadcrumb(B.RowEvent{
			Label: B.BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(len(md.GS.NodeSeedsFn())), X: float64(nodeCount),
		})
	}
}
