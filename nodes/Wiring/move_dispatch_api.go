// move_dispatch_api.go — MoveDispatch's public methods: thin delegators to the owner
// sub-objects composed in move_dispatch.go (mr/gs/sw/ui/lq/rt/scenes/persist), plus the
// two package-private helpers (sendTiltEdit, NodeKind) that read/route across them
// directly. Each owner keeps the actual logic in its own file (mover_registry.go,
// ui_state.go, quantized_move.go, ...); this file exists so the external API surface
// stays discoverable from one place without pulling the construction logic
// (move_dispatch_construct.go) along with it.

package Wiring

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// finalizeActors builds the ring's nodeMover actor directory from md.mr.nodeGeoms, AFTER
// every kind's own build func has run (so every BuildArgs.ClaimSelfDrive call has already
// happened — see moverRegistry.selfDriveClaimed's own doc comment). Thin delegator to
// md.mr (mover_registry.go).
func (md *MoveDispatch) finalizeActors(speedSinks *[]chan float64) {
	md.mr.finalizeActors(speedSinks)
}

// Bind wires the per-edge source Outs and dest wires into each edgeMover. Thin delegator
// to md.mr (mover_registry.go).
func (md *MoveDispatch) Bind(outSink map[string]*wire.Out, slotReg inputcodec.SlotRegistry) {
	md.mr.bind(outSink, slotReg)
}

// Start launches every mover's goroutine. Thin delegator to md.mr (mover_registry.go);
// md.ctx is set here (not part of moverRegistry — see sendMove/enqueueFor's doc
// comments for why sendMove needs it threaded through).
func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	md.ctx = ctx
	return md.mr.start(ctx)
}

// EdgeOut returns the source *Out bound to the given edge label, or nil if unknown.
// Thin delegator to md.mr (mover_registry.go).
func (md *MoveDispatch) EdgeOut(edgeID string) *wire.Out {
	return md.mr.edgeOutFor(edgeID)
}

// centerOfNode returns the current world center for a node id. Thin delegator to md.mr
// (mover_registry.go).
func (md *MoveDispatch) centerOfNode(id string) (vec3, bool) {
	return md.mr.centerOfNode(id)
}

// sendMove routes one moveMsg to a node's dedicated external-entry channel (extIn).
// Thin delegator to md.mr (mover_registry.go); md.ctx is threaded through (not part of
// moverRegistry). This is the bare external-entry path (RootMove, gesture.go) with no
// owning mover goroutine, so it never fires a tap — see nodeMover.tap's doc comment.
func (md *MoveDispatch) sendMove(id string, msg moveMsg) {
	md.mr.sendMove(md.ctx, id, msg)
}

// sendTiltEdit routes one panel-driven tilt-angle click to node id's OWN dedicated
// tiltEditIns channel and returns true, or returns false when id has no such channel (a
// kind that never called BuildArgs.TiltEditIn — every kind except PairNode today),
// telling the caller (applyUpdateTiltVector) to fall back to the old mover-owned path
// instead. Same blocking-with-ctx-cancel-escape shape as sendMove/mr.sendMove, for the
// same reason: this is a bare external-entry send with no owning goroutine to thread a
// ctx from.
func (md *MoveDispatch) sendTiltEdit(id string, msg TiltEditMsg) bool {
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

// enqueueFor returns nm's own non-blocking send function. Thin delegator to md.mr
// (mover_registry.go).
func (md *MoveDispatch) enqueueFor(nm *nodeGeometry) func(id string, msg moveMsg) {
	return md.mr.enqueueFor(nm)
}

// setSelectionUI sets the Go-owned selection (node XOR edge, exclusive). Thin delegator
// to md.ui (ui_state.go).
func (md *MoveDispatch) setSelectionUI(node, edge string) {
	md.ui.setSelectionUI(md.mr.edgeMovers, md.ctx, md.sendMove, node, edge)
}

// setHoverUI sets the Go-owned hover state and MESSAGES the affected node(s). Thin
// delegator to md.ui (ui_state.go).
func (md *MoveDispatch) setHoverUI(node, port string, isInput bool) {
	md.ui.setHoverUI(md.sendMove, node, port, isInput)
}

// NodeKind returns the kind string for the given node id, or "" if unknown.
// Used by applyEdit to resolve the node's kind when snapping a port-anchor
// world-space direction to the nearest ring-anchor index. Called from the
// gesture/stdin-reader goroutine (gesture.go:164, :653), which is NOT the
// nodeMover's own goroutine — this is the ONE genuine cross-goroutine read of
// nm.geom.
//
// Kind lives on nm.geom's embedded nodeIdentity (port_geometry.go), a type carrying
// only the fields the loader sets once at construction and that no handler
// (applyCenter, setPortAnchorId, emitGeometry) ever writes again — grepped clean of
// any write to nodeIdentity's fields outside the load-time literal. That split makes
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

// Quantized scene-polar move math (quantized_move.go): thin delegators to md.lq so
// their existing in-package call sites (tests, move_dispatch_construct.go, gesture.go)
// are unchanged.
func (md *MoveDispatch) heldCenters() map[string]vec3 { return md.lq.heldCenters(md) }
func (md *MoveDispatch) commitNodeMoveLocal(nm *nodeGeometry, newPos vec3) {
	md.lq.commitNodeMoveLocal(md, nm, newPos)
}

// RootMove handles a node-drag under the flat absolute scene-polar layout. Thin
// delegator to md.lq (quantized_move.go).
func (md *MoveDispatch) RootMove(nodeID string, target vec3) bool {
	return md.lq.RootMove(md, nodeID, target)
}

// Overlay-visibility API (MoveDispatch delegators), the overlayState methods, the
// overlayToggles table, defaultOverlayState, and the stdinGuideVisPayload mapper are all
// GENERATED into overlay_gen.go from OVERLAY_FLAG_NAMES (tools/gen-node-defs).
