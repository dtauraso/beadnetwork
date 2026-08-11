// move_dispatch_api.go — MoveDispatch's remaining cross-owner public methods (Start,
// setSelectionUI, NodeKind) plus the package-private helpers (sendMove, sendTiltEdit)
// that read/route across owners directly. Pure single-owner forwards (Bind, EdgeOut,
// centerOfNode, enqueueFor, finalizeActors on md.mr; heldCenters, commitNodeMoveLocal,
// RootMove on md.lq; setHoverUI on md.ui) were deleted — callers address the owner field
// directly (md.mr.X / md.lq.X(md, ...) / md.ui.X). Each owner keeps the actual logic in
// its own file (mover_registry.go, ui_state.go, quantized_move.go, ...).

package Wiring

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

// Start launches every mover's goroutine. Thin delegator to md.mr (mover_registry.go);
// md.ctx is set here (not part of moverRegistry — see sendMove/enqueueFor's doc
// comments for why sendMove needs it threaded through).
func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	md.ctx = ctx
	return md.mr.start(ctx)
}

// sendMove routes one movemsg.Msg to a node's dedicated external-entry channel (extIn).
// Thin delegator to mr (mover_registry.go); ctx is threaded through (not part of
// moverRegistry). This is the bare external-entry path (RootMove, gesture.go) with no
// owning mover goroutine, so it never fires a tap — see nodeMover.tap's doc comment.
func sendMove(mr *moverRegistry, ctx context.Context, id string, msg movemsg.Msg) {
	mr.sendMove(ctx, id, msg)
}

// sendTiltEdit routes one panel-driven tilt-angle click to node id's OWN dedicated
// tiltEditIns channel and returns true, or returns false when id has no such channel (a
// kind that never called BuildArgs.TiltEditIn — every kind except PairNode today),
// telling the caller (applyUpdateTiltVector) to fall back to the old mover-owned path
// instead. Same blocking-with-ctx-cancel-escape shape as sendMove/mr.sendMove, for the
// same reason: this is a bare external-entry send with no owning goroutine to thread a
// ctx from.
func sendTiltEdit(inboxes *nodeInboxes, ctx context.Context, id string, msg movemsg.TiltEditMsg) bool {
	ch, ok := inboxes.tiltEdit[id]
	if !ok {
		return false
	}
	if ctx == nil {
		ch <- msg
		return true
	}
	select {
	case ch <- msg:
	case <-ctx.Done():
	}
	return true
}

// setSelectionUI sets the Go-owned selection (node XOR edge, exclusive). Thin delegator
// to ui (ui_state.go).
func setSelectionUI(ui *uiState, mr *moverRegistry, ctx context.Context, node, edge string) {
	ui.setSelectionUI(mr.edgeMovers, ctx, func(id string, msg movemsg.Msg) { sendMove(mr, ctx, id, msg) }, node, edge)
}

// The overlayState methods, the overlayToggles table, defaultOverlayState, and the
// stdinGuideVisPayload mapper are all GENERATED into overlay_gen.go from
// OVERLAY_FLAG_NAMES (tools/gen-node-defs).
