// ui_state.go — the camera/overlay/gesture/selection/abc-drag UI-state owner split out of
// MoveDispatch (god-object decomposition), as a pure move (no logic changes): uiState
// owns sceneSphere/vp/ov/gest/sel/latchedNode and the
// setSelectionUI/setHoverUI/sendEdgeSelect logic. MoveDispatchs public (and
// package-private) methods of the same names stay as thin delegators, threading through
// the few MoveDispatch fields (edgeMovers/nodeMovers/ctx/sendMove) these handlers need
// but that are NOT part of this extraction.

package Wiring

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/selectionstate"
)

// uiState groups the CURRENTLY-SELECTED (click-select) and CURRENTLY-HOVERED (pointer
// hover) UI-only state (selection_state.go), the polar camera viewpoint (viewpoint_state.go),
// the 9 overlay-toggle visibility booleans (overlay_state.go), the gesture state machine
// (gesture.go), the abc-drag observability bridge, and the scene-sphere reference
// (sphere_layout.go) — pure routing-directory-parked UI state, owned by Go but not part
// of the dispatch/persist/row-table concerns. See each embedded field's own doc comment
// (sceneSphere, vp, ov, gest, sel below) for its own history/reasoning.
type uiState struct {
	// editRefused counts structural edits this run has REFUSED (scene_structure.go). Written
	// only by the view-owner goroutine, which is the only one that handles an edit, and
	// streamed on the Overlay block so the editor can say a gesture did nothing. A counter
	// rather than a flag: the second refusal has to be distinguishable from the first, or
	// making the same mistake twice looks like the editor ignoring you.
	editRefused uint32
	// sceneEditable is SceneTab.Editable for the tree actually loaded, resolved once at load
	// (ResolveSceneDistanceGroups) and streamed on the Overlay block. The palette asks this
	// rather than branching on a scene name in TS.
	sceneEditable bool
	// sceneKinds is the bitmask of kind ids this scene accepts (SceneTab.Kinds), resolved at
	// load beside sceneEditable and streamed with it. Zero until resolved, which reads as
	// "no kind is offered" — the same safe direction sceneEditable takes.
	sceneKinds uint32
	// sceneSphere is the first-class scene reference every node's SCENE polar is measured
	// about (polar-model.md, sphere_layout.go). Loaded from sphere.json (or defaulted from
	// the content-fit) at startup; its Center is the one cartesian anchor. Phase 1 stores
	// it; later phases derive node world from it and move it on pan.
	sceneSphere geom.SceneSphere
	// clockDivisor is this SCENE's ClockDivisor (SceneTab.ClockDivisor, scene/scene_tabs.go),
	// resolved ONCE at load by LoadSpeed from the scene actually loaded (the process is
	// respawned per tab switch, so a per-process value is correct). Defaults to 1 (no
	// scaling) so a bare test-constructed MoveDispatch that never calls LoadSpeed behaves
	// unscaled. clockAttrHandlers's "speed" case and LoadSpeed both feed this through
	// EffectiveClockSpeed (scene_speed_persist.go) so a live slider edit and the load-time
	// seed can never disagree. Never persisted and never crosses the bridge.
	clockDivisor float64
	// hasDistanceGroups is this SCENE's DistanceGroups flag (SceneTab.DistanceGroups,
	// scene/scene_tabs.go), resolved ONCE at load from the scene actually loaded — same
	// per-process shape as clockDivisor above, and correct for the same reason (a tab switch
	// respawns the process). Defaults to FALSE: a bare test-constructed MoveDispatch, and
	// any tree that is not a known scene, must not read the ring's node ids against its own
	// nodes of the same name. Never persisted and never crosses the bridge — what TS sees is
	// three zeroed lengths, which its panel already treats as "no groups".
	hasDistanceGroups bool
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
	// sel groups the CURRENTLY-SELECTED (click-select) and CURRENTLY-HOVERED (pointer hover)
	// UI-only state (selection_state.go) — pure routing-directory-parked UI state, owned by
	// Go but not part of the dispatch/persist/camera concerns. Grouped the same way
	// vp/ov/gest are.
	sel selectionstate.SelectionState

	// --- selection/hover/abc-drag UI state: per-owner, no shared/republished copy ---
	//
	// This state used to live on the old central accumulator only, written by the Trace-drain
	// goroutine on the OTHER end of a round trip from the goroutine that actually sets the
	// intent. It is now owned directly by whichever goroutine sets it: the gesture/
	// MoveDispatch goroutine tracks its OWN local record of the current selection/hover/
	// latched node below (single-owner, mutated only here), and MESSAGES
	// each change to the owning mover's own dedicated channel (movemsg.KindSelect/Hover/
	// Latched — see setSelectionUI/setHoverUI). Each mover stores its
	// OWN selected/hovered/latchedSel/kindID fields (nodeMover) or
	// selected field (edgeMover) and writes them into its own stream frame — no shared map.
	// tr.Select/tr.Hover still fire alongside this, but ONLY for the -trace/.probe EVENT
	// LOG (the central accumulator that used to also feed a fallback packer was deleted
	// entirely — memory/feedback_no_single_writer_bridge.md's final step; WIREFOLD_STREAM_FDS is now
	// mandatory, there is no fallback left).
	//
	// latchedNode is the node id whose LatchedSel bit stays set across a deselect (mirrors
	// the old central accumulator's setSelected latchedSel handling: moves to the newly-selected node, and
	// is left untouched — NOT cleared — on a deselect). Mutated only by the gesture
	// goroutine (setSelectionUI), which also messages the affected movers.
	latchedNode string
	// lastDraggedNode is the id of the most recently DRAG-STARTED node, and (unlike
	// gest.dragNode) it is NEVER cleared back to "" when a drag ends — it only moves
	// when a NEW drag starts. This is what the in-editor "dragging <name>" line reads
	// (via emitViewFrame's dragNodeRow derivation) so that line persists across
	// pointerup instead of vanishing, showing the LAST-dragged node until a different
	// one is dragged. Mutated only by the gesture goroutine, at the same slop-crossing
	// pending→dragging commit edge that sets gest.dragNode (commitDragStart,
	// gesture_graph.go).
	lastDraggedNode string
	// speed is the current playback-speed multiplier (one of the SpeedSlider's six table
	// values: 0, 0.25, 0.5, 0.75, 1, 2). Mirrors, on this goroutine, the value broadcast to
	// every clock-owning goroutine's own speed channel (clockAttrHandlers's "speed" case) —
	// it exists so the VIEW frame can REFLECT the current speed (Buffer.OverlayRow's Speed
	// column) for the webview slider to read, and so LoadSpeed (scene_speed_persist.go) has
	// somewhere to seed the loaded value before the first emit. Defaults to 1 (newMoveDispatch).
	speed float64
	// latticePoints is the pair lattice's current point count (a scene setting, not a
	// per-node one -- every PairNode in a scene runs the same lattice). Seeded ONCE at load
	// by LoadLatticePoints (scene_lattice_persist.go) from view/lattice.json, BEFORE
	// buildNodes runs, so BuildArgs.LatticePointsSeed can hand each node its opening ring
	// size. Defaults to defaultLatticePoints (newMoveDispatch) so a bare test-constructed
	// MoveDispatch that never calls LoadLatticePoints still seeds a valid ring.
	latticePoints int32
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
	msg := movemsg.Msg{Kind: movemsg.KindSelect, Bool: on}
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
func (ui *uiState) setSelectionUI(edgeMovers map[string]*edgeMover, ctx context.Context, sendMove func(id string, msg movemsg.Msg), node, edge string) {
	prevNode := ui.sel.Selected
	prevEdge := ui.sel.SelectedEdge
	ui.sel.Selected = node
	ui.sel.SelectedEdge = edge
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, movemsg.Msg{Kind: movemsg.KindSelect, NodeID: prevNode, Bool: false})
	}
	if node != "" && node != prevNode {
		sendMove(node, movemsg.Msg{Kind: movemsg.KindSelect, NodeID: node, Bool: true})
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
			sendMove(prevLatched, movemsg.Msg{Kind: movemsg.KindLatched, NodeID: prevLatched, Bool: false})
		}
		sendMove(node, movemsg.Msg{Kind: movemsg.KindLatched, NodeID: node, Bool: true})
	}
}

// setHoverUI sets the Go-owned hover state and MESSAGES the affected node(s) to update
// their OWN hovered bit — no shared/republished map. Called only from the gesture
// goroutine (setHover's dedupe check reads ui.sel.HoverNode/Port/Input directly —
// safe since only this same goroutine ever writes them — see gesture.go's
// setHover). sendMove is threaded through from MoveDispatch (not part of uiState).
func (ui *uiState) setHoverUI(sendMove func(id string, msg movemsg.Msg), node, port string, isInput bool) {
	prevNode := ui.sel.HoverNode
	ui.sel.HoverNode, ui.sel.HoverPort, ui.sel.HoverInput = node, port, isInput
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, movemsg.Msg{Kind: movemsg.KindHover, NodeID: prevNode, Bool: false})
	}
	if node != "" {
		sendMove(node, movemsg.Msg{Kind: movemsg.KindHover, NodeID: node, Bool: true, Port: port, IsInput: isInput})
	}
}
