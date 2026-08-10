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
// Thin delegator to md.mr (mover_registry.go); md.ctx is threaded through (not part of
// moverRegistry). This is the bare external-entry path (RootMove, gesture.go) with no
// owning mover goroutine, so it never fires a tap — see nodeMover.tap's doc comment.
func (md *MoveDispatch) sendMove(id string, msg movemsg.Msg) {
	md.mr.sendMove(md.ctx, id, msg)
}

// sendTiltEdit routes one panel-driven tilt-angle click to node id's OWN dedicated
// tiltEditIns channel and returns true, or returns false when id has no such channel (a
// kind that never called BuildArgs.TiltEditIn — every kind except PairNode today),
// telling the caller (applyUpdateTiltVector) to fall back to the old mover-owned path
// instead. Same blocking-with-ctx-cancel-escape shape as sendMove/mr.sendMove, for the
// same reason: this is a bare external-entry send with no owning goroutine to thread a
// ctx from.
func (md *MoveDispatch) sendTiltEdit(id string, msg movemsg.TiltEditMsg) bool {
	ch, ok := md.inboxes.tiltEdit[id]
	if !ok {
		return false
	}
	if md.ctx == nil {
		ch <- msg
		return true
	}
	select {
	case ch <- msg:
	case <-md.ctx.Done():
	}
	return true
}

// setSelectionUI sets the Go-owned selection (node XOR edge, exclusive). Thin delegator
// to md.ui (ui_state.go).
func (md *MoveDispatch) setSelectionUI(node, edge string) {
	md.ui.setSelectionUI(md.mr.edgeMovers, md.ctx, md.sendMove, node, edge)
}

// NodeKind returns the kind string for the given node id, or "" if unknown.
// Used by applyEdit to resolve the node's kind when snapping a port-anchor
// world-space direction to the nearest ring-anchor index. Called from the
// gesture/stdin-reader goroutine (gesture.go:164, :653), which is NOT the
// nodeMover's own goroutine — this is the ONE genuine cross-goroutine read of
// nm.geom.
//
// Kind lives on nm.geom's embedded nodegeom.NodeIdentity (nodegeom/port_geometry.go), a type carrying
// only the fields the loader sets once at construction and that no handler
// (applyCenter, setPortAnchorId, emitGeometry) ever writes again — grepped clean of
// any write to NodeIdentity's fields outside the load-time literal. That split makes
// this safe by CONSTRUCTION rather than by coincidence of which byte ranges a
// particular access happens to touch: identity fields are not merely
// unwritten-in-practice today, they are not reachable from any writer's
// field-assignment at all, in a different embedded struct from the mutable
// ScenePolar/HasPos/ReachR/Inputs/Outputs applyCenter and setPortAnchorId do write.
// TestNodeKindConcurrentWithApplyCenterUnderRace exercises this concurrently under
// -race as a regression check on the split holding.
func (md *MoveDispatch) NodeKind(nodeID string) string {
	if nm, ok := md.mr.nodeGeoms[nodeID]; ok {
		return nm.geom.Kind
	}
	return ""
}

// The overlayState methods, the overlayToggles table, defaultOverlayState, and the
// stdinGuideVisPayload mapper are all GENERATED into overlay_gen.go from
// OVERLAY_FLAG_NAMES (tools/gen-node-defs).
