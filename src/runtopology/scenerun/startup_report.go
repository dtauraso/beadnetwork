package scenerun

import (
	T "github.com/dtauraso/wirefold/src/Trace"
)

func EmitStartupBreadcrumbs(md *MoveDispatch, scenePath string, nodeCount int) {

	md.UI.EmitBreadcrumb(T.RowEvent{
		Label: T.BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(nodeCount), Text: scenePath,
	})
}

func CheckRowSeedCount(md *MoveDispatch, nodeCount int) {
	if len(md.GS.NodeSeedsFn()) != nodeCount {

		md.UI.EmitBreadcrumb(T.RowEvent{
			Label: T.BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(len(md.GS.NodeSeedsFn())), X: float64(nodeCount),
		})
	}
}
