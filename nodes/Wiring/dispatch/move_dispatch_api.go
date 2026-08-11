// move_dispatch_api.go — MoveDispatch's remaining cross-owner public methods (Start,
// setSelectionUI, NodeKind) plus the package-private helpers (sendMove, sendTiltEdit)
// that read/route across owners directly. Pure single-owner forwards (Bind, EdgeOut,
// centerOfNode, enqueueFor, finalizeActors on md.MR; heldCenters, commitNodeMoveLocal,
// RootMove on md.LQ; setHoverUI on md.ui) were deleted — callers address the owner field
// directly (md.MR.X / md.LQ.X(md, ...) / md.ui.X). Each owner keeps the actual logic in
// its own file (mover_registry.go, ui_state.go, quantized_move.go, ...).

package dispatch

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeinbox"
)

// Start launches every mover's goroutine. Thin delegator to md.MR (mover_registry.go);
// md.ctx is set here (not part of moverRegistry — see sendMove/enqueueFor's doc
// comments for why sendMove needs it threaded through).
func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	md.ctx = ctx
	return md.MR.Start(ctx)
}

// SendMove routes one movemsg.Msg to a node's dedicated external-entry channel (extIn).
// Thin delegator to mr (nodes/Wiring/moverreg); ctx is threaded through (not part of
// moverreg.MoverRegistry). This is the bare external-entry path (RootMove, gesture.go) with
// no owning mover goroutine, so it never fires a tap — see nodeMover.tap's doc comment.
// Exported (§30, docs/planning/movedispatch-decomposition.md) so the stdin cluster
// (nodes/Wiring/stdinreader) can call it without reaching back into this package for an
// unexported name.
func SendMove(mr *moverreg.MoverRegistry, ctx context.Context, id string, msg movemsg.Msg) {
	mr.SendMove(ctx, id, msg)
}

// SendTiltEdit routes one panel-driven tilt-angle click to node id's OWN dedicated
// tiltEditIns channel and returns true, or returns false when id has no such channel (a
// kind that never called BuildArgs.TiltEditIn — every kind except PairNode today),
// telling the caller (applyUpdateTiltVector) to fall back to the old mover-owned path
// instead. Thin delegator to nodeinbox.NodeInboxes.SendTiltEdit (nodes/Wiring/nodeinbox,
// lifted out of this package in docs/planning/movedispatch-decomposition.md §29). Exported
// (§30) for the same reason as SendMove.
func SendTiltEdit(inboxes *nodeinbox.NodeInboxes, ctx context.Context, id string, msg movemsg.TiltEditMsg) bool {
	return inboxes.SendTiltEdit(ctx, id, msg)
}

// setSelectionUI and sendEdgeSelect moved to nodes/Wiring/gesture (§31,
// docs/planning/movedispatch-decomposition.md) — they were used ONLY by the gesture
// cluster's applySelect, which moved with them.

// NodeSelfDriven reports whether node id's own geometry is driven by that node's own kind
// goroutine (task/pair-node-owns-itself, ClaimSelfDrive) rather than a separate NodeMover
// goroutine. Exposed for verification: the model's whole point — one goroutine, not two,
// for the same node id — is otherwise invisible from outside this package (package main's
// own headless tests are the only place every kind, PairNode included, is registered — see
// kind_registry_parity_test.go's own doc comment). Thin delegator to md.MR
// (mover_registry.go); kept on MoveDispatch because package main's own tests call it and
// mr is unexported. Moved here from the former pair_node_self.go in §20
// (docs/planning/movedispatch-decomposition.md) when PairNodeSelf itself moved to package
// nodeactor — this method stays in package Wiring because MoveDispatch is a Wiring type.
func (md *MoveDispatch) NodeSelfDriven(id string) bool {
	return md.MR.NodeSelfDriven(id)
}

// HasNodeMover reports whether node id has a real, separate NodeMover actor (a ring
// node) as opposed to no NodeMover at all (a self-driven pair node, or an unknown id).
// Thin delegator to md.MR, kept for the same reason as NodeSelfDriven.
func (md *MoveDispatch) HasNodeMover(id string) bool {
	return md.MR.HasNodeMover(id)
}

// NodeQuantOffset returns node id's own current quantized polar offset triple
// (iTheta, iPhi, iR), for the same external-verification reason as NodeSelfDriven — e.g.
// confirming a real reload lands on the same offset a live edit just persisted. Thin
// delegator to md.MR, kept for the same reason as NodeSelfDriven.
func (md *MoveDispatch) NodeQuantOffset(id string) (iTheta, iPhi, iR int, ok bool) {
	return md.MR.NodeQuantOffset(id)
}

// The OverlayState methods, the OverlayToggles table, DefaultOverlayState, and the
// stdinGuideVisPayload mapper are all GENERATED into nodes/Wiring/viewstate/overlay_state.go
// from OVERLAY_FLAG_NAMES (tools/gen-node-defs).
