package runtopology

import (
	"fmt"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
)

// emitStartupBreadcrumbs announces which scene loaded, on both the breadcrumb channel and
// the VIEW stream. nodeCount is len(nodes) from LoadTopology.
func emitStartupBreadcrumbs(tr *T.Trace, md *W.MoveDispatch, scenePath string, nodeCount int) {
	// One example startup breadcrumb — proves the debug channel end-to-end and is genuinely
	// useful (which topology loaded, how many nodes). Sparse: once per run.
	tr.Breadcrumb("topology-loaded", scenePath, "", fmt.Sprintf("nodes=%d", nodeCount))
	// Structured buffer counterpart: rides the VIEW stream (no per-node stream exists
	// yet for a startup-only event, and this runs on the main goroutine before any
	// per-node/edge/interior goroutine exists). topologyPath is genuinely free-form
	// (a filesystem path), so it rides the sanctioned Text column; nodes count is
	// the typed Value column.
	md.UI.EmitBreadcrumb(wire.RowEvent{
		Label: T.BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(nodeCount), Text: scenePath,
	})
}

// checkRowSeedCount is the row-seed sanity check. nodeCount is len(nodes).
//
// Sparse, one-time startup sanity check (CLAUDE.md DEBUG BREADCRUMB channel): every
// node LoadTopology returned should have a row-seed entry (md.GS.NodeSeedsFn(), the SAME
// spec-order row table nodes/Wiring's own move-dispatch/stream wiring above already
// uses). A mismatch means md.GS.NodeSeedsFn() (spec order) and LoadTopology's node list
// diverged — a real topology bug — and must be visible.
func checkRowSeedCount(tr *T.Trace, md *W.MoveDispatch, nodeCount int) {
	if len(md.GS.NodeSeedsFn()) != nodeCount {
		tr.Breadcrumb("row-seed-count-mismatch", "", "", fmt.Sprintf("NodeSeeds=%d nodes=%d", len(md.GS.NodeSeedsFn()), nodeCount))
		// Structured buffer counterpart, VIEW stream (same reasoning as
		// topology-loaded above). Value=NodeSeeds count, X=nodes count — both
		// small typed ints, no free-form text needed.
		md.UI.EmitBreadcrumb(wire.RowEvent{
			Label: T.BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(len(md.GS.NodeSeedsFn())), X: float64(nodeCount),
		})
	}
}
