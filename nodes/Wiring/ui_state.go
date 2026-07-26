// ui_state.go — the camera/overlay/gesture/selection/abc-drag UI-state owner split out of
// MoveDispatch (god-object decomposition), as a pure move (no logic changes): uiState
// owns sceneSphere/vp/ov/gest/sel/abcDragCh/abcDragCount/latchedNode and the
// setSelectionUI/setHoverUI/sendEdgeSelect/resetAbcDrag logic. MoveDispatch's public (and
// package-private) methods of the same names stay as thin delegators, threading through
// the few MoveDispatch fields (edgeMovers/nodeMovers/ctx/sendMove) these handlers need
// but that are NOT part of this extraction.

package Wiring

import (
	"context"
)

// uiState groups the CURRENTLY-SELECTED (click-select) and CURRENTLY-HOVERED (pointer
// hover) UI-only state (selection_state.go), the polar camera viewpoint (viewpoint_state.go),
// the 9 overlay-toggle visibility booleans (overlay_state.go), the gesture state machine
// (gesture.go), the abc-drag observability bridge, and the scene-sphere reference
// (sphere_layout.go) — pure routing-directory-parked UI state, owned by Go but not part
// of the dispatch/persist/row-table concerns. See each embedded field's own doc comment
// (sceneSphere, vp, ov, gest, sel below) for its own history/reasoning.
type uiState struct {
	// sceneSphere is the first-class scene reference every node's SCENE polar is measured
	// about (polar-model.md, sphere_layout.go). Loaded from scene.json (or defaulted from
	// the content-fit) at startup; its Center is the one cartesian anchor. Phase 1 stores
	// it; later phases derive node world from it and move it on pan.
	sceneSphere sceneSphere
	// vp is the polar camera viewpoint state (viewpoint_state.go). Owned entirely by
	// MoveDispatch — no separate goroutine; callers serialize externally (stdin reader
	// runs in a single goroutine). MoveDispatch exposes thin delegating methods.
	vp viewpointState
	// ov groups the 9 overlay-toggle visibility booleans and their flip/emit logic
	// (overlay_state.go). Initialized to defaults by newMoveDispatch (all true).
	// MoveDispatch exposes thin delegating methods.
	ov overlayState
	// gest is the gesture state machine (gesture.go): it consumes raw pointer/wheel input
	// and produces camera (viewpoint) + topology (node-move) changes. Owned by
	// MoveDispatch; serialized by the single-goroutine stdin reader. Zero value = idle.
	gest gestureState
	// abcDragCh is the non-blocking, message-passing bridge from an abc-drag RECIPIENT's
	// own nodeMover goroutine (quantized_move.go's neighborSetCRequantize, potentially many
	// different goroutines) to the ONE gesture/stdin-reader goroutine that owns
	// abcDragCount and writes the VIEW frame. Per MODEL.md's no-shared-state directive:
	// message-passing, never a shared counter. A full channel just drops that one
	// count-observability tick (no delivery guarantee, same shape as every other
	// fire-and-forget bridge in this codebase) rather than blocking the recipient's own
	// goroutine. nil until SetViewStream runs (no dedicated view stream ⇒ nothing to send
	// to; nodeMover call sites nil-check before sending).
	abcDragCh chan struct{}
	// abcDragCount is this goroutine's OWN plain int (only the
	// gesture/stdin-reader goroutine ever reads or writes it, via DrainAbcDragChan) — the
	// VIEW frame's Overlay.AbcDragCount column reads this directly. Per-drag: reset to 0
	// at the start of each drag via resetAbcDrag (not cumulative for the run's lifetime).
	abcDragCount uint32
	// sel groups the CURRENTLY-SELECTED (click-select) and CURRENTLY-HOVERED (pointer hover)
	// UI-only state (selection_state.go) — pure routing-directory-parked UI state, owned by
	// Go but not part of the dispatch/persist/camera concerns. Grouped the same way
	// vp/ov/gest are.
	sel selectionState

	// --- selection/hover/abc-drag UI state: per-owner, no shared/republished copy ---
	//
	// This state used to live on the old central accumulator only, written by the Trace-drain
	// goroutine on the OTHER end of a round trip from the goroutine that actually sets the
	// intent. It is now owned directly by whichever goroutine sets it: the gesture/
	// MoveDispatch goroutine tracks its OWN local record of the current selection/hover/
	// latched node below (single-owner, mutated only here), and MESSAGES
	// each change to the owning mover's own dedicated channel (moveMsgKindSelect/Hover/
	// Latched/AbcReset — see setSelectionUI/setHoverUI/resetAbcDrag). Each mover stores its
	// OWN selected/hovered/latchedSel/gotDragMsg/dragDelta*/kindID fields (nodeMover) or
	// selected field (edgeMover) and writes them into its own stream frame — no shared map.
	// tr.Select/tr.Hover/tr.AbcDrag/tr.AbcDragReset still fire
	// alongside this, but ONLY for the -trace/.probe EVENT LOG (the central accumulator
	// that used to also feed a fallback packer was deleted
	// entirely — memory/feedback_no_single_writer_bridge.md's final step; WIREFOLD_STREAM_FDS is now
	// mandatory, there is no fallback left).
	//
	// latchedNode is the node id whose LatchedSel bit stays set across a deselect (mirrors
	// the old central accumulator's setSelected latchedSel handling: moves to the newly-selected node, and
	// is left untouched — NOT cleared — on a deselect). Mutated only by the gesture
	// goroutine (setSelectionUI), which also messages the affected movers.
	latchedNode string
}

// sendEdgeSelect routes a select/deselect message to one edge's OWN dedicated extIn
// channel (mirrors sendMove's node counterpart) — the edgeMover sets its OWN selected
// field on its own goroutine, no shared map. A blocking send with a ctx-cancel escape
// hatch, same reasoning as sendMove. edgeMovers/ctx are threaded through from
// MoveDispatch (not part of uiState).
func (ui *uiState) sendEdgeSelect(edgeMovers map[string]*edgeMover, ctx context.Context, label string, on bool) {
	em, ok := edgeMovers[label]
	if !ok {
		return
	}
	msg := moveMsg{Kind: moveMsgKindSelect, Bool: on}
	if ctx == nil {
		em.extIn <- msg
		return
	}
	select {
	case em.extIn <- msg:
	case <-ctx.Done():
	}
}

// setSelectionUI sets the Go-owned selection (node XOR edge, exclusive — mirrors
// the old central accumulator's setSelected/setSelectedEdge exclusivity), moving latchedNode to
// a newly-selected node (left untouched on a deselect), and MESSAGES every affected
// mover to update its OWN selected/latchedSel bit — no shared/republished map. Called
// only from the gesture/MoveDispatch goroutine (applySelect); ui.sel/ui.latchedNode are
// mutated only here, and each message ride the mover's own
// dedicated channel so the mover mutates only its own fields on its own goroutine.
// edgeMovers/ctx/sendMove are threaded through from MoveDispatch (not part of uiState).
func (ui *uiState) setSelectionUI(edgeMovers map[string]*edgeMover, ctx context.Context, sendMove func(id string, msg moveMsg), node, edge string) {
	prevNode := ui.sel.selected
	prevEdge := ui.sel.selectedEdge
	ui.sel.selected = node
	ui.sel.selectedEdge = edge
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, moveMsg{Kind: moveMsgKindSelect, NodeID: prevNode, Bool: false})
	}
	if node != "" && node != prevNode {
		sendMove(node, moveMsg{Kind: moveMsgKindSelect, NodeID: node, Bool: true})
	}
	if prevEdge != "" && prevEdge != edge {
		ui.sendEdgeSelect(edgeMovers, ctx, prevEdge, false)
	}
	if edge != "" && edge != prevEdge {
		ui.sendEdgeSelect(edgeMovers, ctx, edge, true)
	}
	if node != "" && node != ui.latchedNode {
		prevLatched := ui.latchedNode
		ui.latchedNode = node
		if prevLatched != "" {
			sendMove(prevLatched, moveMsg{Kind: moveMsgKindLatched, NodeID: prevLatched, Bool: false})
		}
		sendMove(node, moveMsg{Kind: moveMsgKindLatched, NodeID: node, Bool: true})
	}
}

// setHoverUI sets the Go-owned hover state and MESSAGES the affected node(s) to update
// their OWN hovered bit — no shared/republished map. Called only from the gesture
// goroutine (setHover's dedupe check reads ui.sel.hoverNode/Port/Input directly —
// safe since only this same goroutine ever writes them — see gesture.go's
// setHover). sendMove is threaded through from MoveDispatch (not part of uiState).
func (ui *uiState) setHoverUI(sendMove func(id string, msg moveMsg), node, port string, isInput bool) {
	prevNode := ui.sel.hoverNode
	ui.sel.hoverNode, ui.sel.hoverPort, ui.sel.hoverInput = node, port, isInput
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, moveMsg{Kind: moveMsgKindHover, NodeID: prevNode, Bool: false})
	}
	if node != "" {
		sendMove(node, moveMsg{Kind: moveMsgKindHover, NodeID: node, Bool: true, Port: port, IsInput: isInput})
	}
}

// resetAbcDrag re-scopes the recipient SET to the drag about to start: MESSAGES every
// node mover to clear its OWN abc-drag recipient bit (mirrors the old central accumulator's
// KindAbcDragReset handling), and zeroes ui.abcDragCount so the "drag received ×{count}"
// counter is per-drag rather than cumulative for the run's lifetime. Called from the
// gesture goroutine at the pending→dragging transition (gesture.go). The recipient-bit
// broadcast is not a shared flag: each mover clears its own bit on its own goroutine, no
// generation counter. The abcDragCount write is a plain field write because this method
// runs on the same (gesture/stdin-reader) goroutine that owns abcDragCount. nodeMovers/
// sendMove are threaded through from MoveDispatch (not part of uiState).
func (ui *uiState) resetAbcDrag(nodeMovers map[string]*nodeMover, sendMove func(id string, msg moveMsg)) {
	ui.abcDragCount = 0
	for id := range nodeMovers {
		sendMove(id, moveMsg{Kind: moveMsgKindAbcReset, NodeID: id})
	}
}
