// move_dispatch_api.go — MoveDispatch's remaining cross-owner public methods (Start,
// setSelectionUI, NodeKind) plus the package-private helpers (sendMove, sendTiltEdit)
// that read/route across owners directly. Pure single-owner forwards (Bind, EdgeOut,
// centerOfNode, enqueueFor, finalizeActors on md.mr; heldCenters, commitNodeMoveLocal,
// RootMove on md.lq; setHoverUI on md.ui) were deleted — callers address the owner field
// directly (md.mr.X / md.lq.X(md, ...) / md.ui.X). Each owner keeps the actual logic in
// its own file (mover_registry.go, ui_state.go, quantized_move.go, ...).

package dispatch

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// Start launches every mover's goroutine. Thin delegator to md.mr (mover_registry.go);
// md.ctx is set here (not part of moverRegistry — see sendMove/enqueueFor's doc
// comments for why sendMove needs it threaded through).
func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	md.ctx = ctx
	return md.mr.Start(ctx)
}

// sendMove routes one movemsg.Msg to a node's dedicated external-entry channel (extIn).
// Thin delegator to mr (nodes/Wiring/moverreg); ctx is threaded through (not part of
// moverreg.MoverRegistry). This is the bare external-entry path (RootMove, gesture.go) with
// no owning mover goroutine, so it never fires a tap — see nodeMover.tap's doc comment.
func sendMove(mr *moverreg.MoverRegistry, ctx context.Context, id string, msg movemsg.Msg) {
	mr.SendMove(ctx, id, msg)
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

// setSelectionUI sets the Go-owned selection (node XOR edge, exclusive). Thin delegator to
// ui.SetSelectionUI (nodes/Wiring/viewstate), binding the two closures it needs in place of
// the *moverreg.MoverRegistry/edgeMover-map parameters that type cannot name — same
// bound-func treatment as the gesture cluster
// (docs/planning/movedispatch-decomposition.md section 6).
func setSelectionUI(ui *viewstate.UIState, mr *moverreg.MoverRegistry, ctx context.Context, node, edge string) {
	sendMoveFn := func(id string, msg movemsg.Msg) { sendMove(mr, ctx, id, msg) }
	sendEdgeSelectFn := func(label string, on bool) { sendEdgeSelect(mr.EdgeMovers(), ctx, label, on) }
	ui.SetSelectionUI(sendMoveFn, sendEdgeSelectFn, node, edge)
}

// NodeSelfDriven reports whether node id's own geometry is driven by that node's own kind
// goroutine (task/pair-node-owns-itself, ClaimSelfDrive) rather than a separate NodeMover
// goroutine. Exposed for verification: the model's whole point — one goroutine, not two,
// for the same node id — is otherwise invisible from outside this package (package main's
// own headless tests are the only place every kind, PairNode included, is registered — see
// kind_registry_parity_test.go's own doc comment). Thin delegator to md.mr
// (mover_registry.go); kept on MoveDispatch because package main's own tests call it and
// mr is unexported. Moved here from the former pair_node_self.go in §20
// (docs/planning/movedispatch-decomposition.md) when PairNodeSelf itself moved to package
// nodeactor — this method stays in package Wiring because MoveDispatch is a Wiring type.
func (md *MoveDispatch) NodeSelfDriven(id string) bool {
	return md.mr.NodeSelfDriven(id)
}

// HasNodeMover reports whether node id has a real, separate NodeMover actor (a ring
// node) as opposed to no NodeMover at all (a self-driven pair node, or an unknown id).
// Thin delegator to md.mr, kept for the same reason as NodeSelfDriven.
func (md *MoveDispatch) HasNodeMover(id string) bool {
	return md.mr.HasNodeMover(id)
}

// NodeQuantOffset returns node id's own current quantized polar offset triple
// (iTheta, iPhi, iR), for the same external-verification reason as NodeSelfDriven — e.g.
// confirming a real reload lands on the same offset a live edit just persisted. Thin
// delegator to md.mr, kept for the same reason as NodeSelfDriven.
func (md *MoveDispatch) NodeQuantOffset(id string) (iTheta, iPhi, iR int, ok bool) {
	return md.mr.NodeQuantOffset(id)
}

// sendEdgeSelect routes a select/deselect message to one edge's OWN dedicated extIn
// channel (mirrors sendMove's node counterpart) — the edgeMover sets its OWN selected
// field on its own goroutine, no shared map. A blocking send with a ctx-cancel escape
// hatch, same reasoning as sendMove. Moved from viewstate/ui_state.go's old
// (*uiState).sendEdgeSelect (docs/planning/gesture-actor.md's lift): it needs the
// edgeMover map, an unexported Wiring type viewstate cannot name, so it stays here and is
// handed to UIState.SetSelectionUI as a bound func value (setSelectionUI above).
func sendEdgeSelect(edgeMovers map[string]*edgemover.EdgeMover, ctx context.Context, label string, on bool) {
	em, ok := edgeMovers[label]
	if !ok {
		return
	}
	em.Select(ctx, on)
}

// The OverlayState methods, the OverlayToggles table, DefaultOverlayState, and the
// stdinGuideVisPayload mapper are all GENERATED into nodes/Wiring/viewstate/overlay_state.go
// from OVERLAY_FLAG_NAMES (tools/gen-node-defs).
