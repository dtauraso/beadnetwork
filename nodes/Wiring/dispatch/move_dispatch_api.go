// move_dispatch_api.go — MoveDispatch's one remaining cross-owner public method, Start.
// SendMove/SendTiltEdit/NodeSelfDriven/HasNodeMover/NodeQuantOffset were pure single-owner
// forwards onto md.MR/md.Inboxes and were deleted (docs/planning/movedispatch-decomposition.md,
// the remainder cluster) — every caller now addresses the owner field directly
// (md.MR.SendMove/md.Inboxes.SendTiltEdit/md.MR.NodeSelfDriven/md.MR.HasNodeMover/
// md.MR.NodeQuantOffset). setSelectionUI and sendEdgeSelect moved to nodes/Wiring/gesture
// (§31) earlier. Each owner keeps the actual logic in its own file (mover_registry.go,
// node_inbox.go, ...).

package dispatch

import (
	"context"
	"sync"
)

// Start launches every mover's goroutine. Thin delegator to md.MR (mover_registry.go);
// md.ctx is set here (not part of moverRegistry — see sendMove/enqueueFor's doc
// comments for why sendMove needs it threaded through).
func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	md.ctx = ctx
	return md.MR.Start(ctx)
}

// The OverlayState methods, the OverlayToggles table, DefaultOverlayState, and the
// stdinGuideVisPayload mapper are all GENERATED into nodes/Wiring/viewstate/overlay_state.go
// from OVERLAY_FLAG_NAMES (tools/gen-node-defs).
