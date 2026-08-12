package runtopology

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
)

func emitStartupBreadcrumbs(tr *T.Trace, md *W.MoveDispatch, scenePath string, nodeCount int) {

	tr.Breadcrumb("topology-loaded", scenePath, "", fmt.Sprintf("nodes=%d", nodeCount))

	md.UI.EmitBreadcrumb(rowevent.RowEvent{
		Label: T.BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(nodeCount), Text: scenePath,
	})
}

func checkRowSeedCount(tr *T.Trace, md *W.MoveDispatch, nodeCount int) {
	if len(md.GS.NodeSeedsFn()) != nodeCount {
		tr.Breadcrumb("row-seed-count-mismatch", "", "", fmt.Sprintf("NodeSeeds=%d nodes=%d", len(md.GS.NodeSeedsFn()), nodeCount))

		md.UI.EmitBreadcrumb(rowevent.RowEvent{
			Label: T.BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(len(md.GS.NodeSeedsFn())), X: float64(nodeCount),
		})
	}
}
