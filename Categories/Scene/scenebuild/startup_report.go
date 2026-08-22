package scenebuild

import (
	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

func EmitStartupBreadcrumbs(md *scenerun.MoveDispatch, scenePath string, nodeCount int) {

	md.UI.EmitBreadcrumb(viewstate.RowEvent{
		Label: BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(nodeCount), Text: scenePath,
	})
}

func CheckRowSeedCount(md *scenerun.MoveDispatch, nodeCount int) {
	if len(md.GS.NodeSeedsFn()) != nodeCount {

		md.UI.EmitBreadcrumb(viewstate.RowEvent{
			Label: BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(len(md.GS.NodeSeedsFn())), X: float64(nodeCount),
		})
	}
}
