// ui_state.go — the camera/overlay/gesture/selection/abc-drag UI-state owner lifted out of
// MoveDispatch (god-object decomposition, docs/planning/movedispatch-decomposition.md
// section 6/6b/6c, and the lift completed here per docs/planning/gesture-actor.md): UIState
// owns sceneSphere/vp/ov/gest/sel/latchedNode, the VIEW stream's own write side
// (view_stream.go in this package), and the selection/hover/drag-plane leaf logic that used
// to live on Wiring's unexported uiState. MoveDispatch holds this type directly as an
// exported field (UI viewstate.UIState) — the same "export follows the type moving" pattern
// already used for GS/RT/Scenes.
//
// Two owner-typed dependencies (moverRegistry, layoutQuantizer, and the edgeMover directory)
// cannot be named here — they are unexported types in package Wiring. Every method that used
// to need one takes a BOUND FUNC VALUE instead (sendMove, sendEdgeSelect, NodeRowFor,
// DistanceGroupLensFn), the same treatment already applied to the gesture cluster's
// beginSphereRotation/applyNodeDragTarget/commitDragStart/setHover
// (docs/planning/movedispatch-decomposition.md section 6). NodeRowFor/DistanceGroupLensFn are
// bound ONCE at construction (move_dispatch_construct.go), mirroring
// ng.msg.sendMove = md.mr.enqueueFor(ng) — not threaded through every call.
package viewstate

import (
	"fmt"
	"math"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/selectionstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// UIState groups the CURRENTLY-SELECTED (click-select) and CURRENTLY-HOVERED (pointer
// hover) UI-only state, the polar camera viewpoint, the overlay-toggle visibility booleans,
// the gesture state machine, the abc-drag observability bridge, the scene-sphere reference,
// and the VIEW stream's own write side — pure routing-directory-parked UI state, owned by Go
// but not part of the dispatch/persist/row-table concerns. See each field's own doc comment
// for its own history/reasoning (carried over from nodes/Wiring/ui_state.go verbatim where
// unchanged).
type UIState struct {
	// EditRefused counts structural edits this run has REFUSED (Wiring's scene_structure.go).
	// Written only by the view-owner goroutine, which is the only one that handles an edit,
	// and streamed on the Overlay block so the editor can say a gesture did nothing. A
	// counter rather than a flag: the second refusal has to be distinguishable from the
	// first, or making the same mistake twice looks like the editor ignoring you.
	EditRefused uint32
	// SceneEditable is SceneTab.Editable for the tree actually loaded, resolved once at load
	// (Wiring's ResolveSceneDistanceGroups) and streamed on the Overlay block.
	SceneEditable bool
	// SceneKinds is the bitmask of kind ids this scene accepts (SceneTab.Kinds), resolved at
	// load beside SceneEditable and streamed with it. Zero until resolved.
	SceneKinds uint32
	// SceneSphere is the first-class scene reference every node's SCENE polar is measured
	// about (polar-model.md). Loaded from sphere.json (or defaulted from the content-fit) at
	// startup; its Center is the one cartesian anchor.
	SceneSphere geom.SceneSphere
	// ClockDivisor is this SCENE's ClockDivisor (SceneTab.ClockDivisor, scene/scene_tabs.go),
	// resolved ONCE at load by Wiring's LoadSpeed. Defaults to 1 (no scaling) so a bare
	// test-constructed UIState behaves unscaled. Never persisted and never crosses the
	// bridge.
	ClockDivisor float64
	// HasDistanceGroups is this SCENE's DistanceGroups flag (SceneTab.DistanceGroups,
	// scene/scene_tabs.go), resolved ONCE at load. Defaults to FALSE. Never persisted and
	// never crosses the bridge.
	HasDistanceGroups bool
	// VP is the polar camera viewpoint state (nodes/Wiring/gesturefsm.ViewpointState). Owned
	// entirely by the view-owner goroutine; callers serialize externally.
	VP gesturefsm.ViewpointState
	// OV groups the overlay-toggle visibility booleans and their flip/emit logic
	// (overlay_state.go, this package). Initialized to defaults by DefaultOverlayState.
	OV OverlayState
	// Gest is the gesture state machine (gesturefsm.GestureState): it consumes raw
	// pointer/wheel input and produces camera (viewpoint) + topology (node-move) changes.
	// Zero value = idle.
	Gest gesturefsm.GestureState
	// Sel groups the CURRENTLY-SELECTED (click-select) and CURRENTLY-HOVERED (pointer hover)
	// UI-only state (nodes/Wiring/selectionstate).
	Sel selectionstate.SelectionState

	// LatchedNode is the node id whose LatchedSel bit stays set across a deselect (moves to
	// the newly-selected node, and is left untouched — NOT cleared — on a deselect). Mutated
	// only by the gesture goroutine (SetSelectionUI), which also messages the affected
	// movers.
	LatchedNode string
	// LastDraggedNode is the id of the most recently DRAG-STARTED node, and (unlike
	// Gest.DragNode) it is NEVER cleared back to "" when a drag ends — it only moves when a
	// NEW drag starts. This is what the in-editor "dragging <name>" line reads (via
	// EmitViewFrame's dragNodeRow derivation) so that line persists across pointerup instead
	// of vanishing.
	LastDraggedNode string
	// Speed is the current playback-speed multiplier (one of the SpeedSlider's six table
	// values: 0, 0.25, 0.5, 0.75, 1, 2). Defaults to 1.
	Speed float64
	// LatticePoints is the pair lattice's current point count (a scene setting, not a
	// per-node one). Seeded ONCE at load, BEFORE buildNodes runs. Defaults to a caller-
	// supplied value (Wiring's move_dispatch_construct.go) so a bare test-constructed
	// UIState still seeds a valid ring.
	LatticePoints int32

	// NodeRowFor resolves a node id to its buffer row (bound once, at construction, to
	// md.RT.NodeRowFor — nodes/Wiring/rowtables.RowTables is exported, but UIState cannot
	// hold a *rowtables.RowTables field of its own without MoveDispatch handing it one; a
	// bound func value is the same treatment the gesture cluster already uses for mr/lq).
	// nil-checked by EmitViewFrame (a bare test-constructed UIState never sets this).
	NodeRowFor func(id string) (int32, bool)
	// DistanceGroupLensFn returns the 3 distance-group panel lengths (bound once, at
	// construction, to a closure over Wiring's DistanceGroupLens(ui, mr) — DistanceGroupLens
	// itself stays in Wiring because it reads *moverRegistry, an unexported type this
	// package cannot name). nil-checked by EmitViewFrame.
	DistanceGroupLensFn func() (timeLen, inputLen, gateLen float32)

	// --- the dedicated VIEW stream's own write side (view_stream.go, this package) ---
	viewOut        viewClaimedStream
	ViewBuildFrame ViewFrameBuilder
	viewTick       uint32
	viewClaimed    bool
}

// SetSelectionUI sets the Go-owned selection (node XOR edge, exclusive), moving LatchedNode
// to a newly-selected node (left untouched on a deselect), and MESSAGES every affected mover
// to update its OWN selected/latchedSel bit — no shared/republished map. sendMove/
// sendEdgeSelect are bound func values closing over the caller's own moverRegistry/
// edgeMover directory (Wiring types this package cannot name) — see
// Wiring/move_dispatch_api.go's setSelectionUI wrapper, the caller.
func (ui *UIState) SetSelectionUI(sendMove func(id string, msg movemsg.Msg), sendEdgeSelect func(label string, on bool), node, edge string) {
	prevNode := ui.Sel.Selected
	prevEdge := ui.Sel.SelectedEdge
	ui.Sel.Selected = node
	ui.Sel.SelectedEdge = edge
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, movemsg.Msg{Kind: movemsg.KindSelect, NodeID: prevNode, Bool: false})
	}
	if node != "" && node != prevNode {
		sendMove(node, movemsg.Msg{Kind: movemsg.KindSelect, NodeID: node, Bool: true})
	}
	if prevEdge != "" && prevEdge != edge {
		sendEdgeSelect(prevEdge, false)
	}
	if edge != "" && edge != prevEdge {
		sendEdgeSelect(edge, true)
	}
	if node != "" && node != ui.LatchedNode {
		prevLatched := ui.LatchedNode
		ui.LatchedNode = node
		if prevLatched != "" {
			sendMove(prevLatched, movemsg.Msg{Kind: movemsg.KindLatched, NodeID: prevLatched, Bool: false})
		}
		sendMove(node, movemsg.Msg{Kind: movemsg.KindLatched, NodeID: node, Bool: true})
	}
}

// DropPointFromNDC unprojects a drop's screen position onto the camera-facing plane through
// the SCENE CENTRE — the same ray-through-NDC a node drag already unprojects (DragPlaneHit
// below), against a plane that exists whether or not anything was under the pointer.
// ok=false when the ray is parallel to the plane or the hit is non-finite.
func (ui *UIState) DropPointFromNDC(ndcX, ndcY float64) (wire.Vec3, bool) {
	vp := ui.VP.Viewpoint
	eye := geom.EyeOf(vp)
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	dir := geom.RayDirThroughNDC(ndcX, ndcY, basis, ui.Gest.Fov, ui.Gest.Rect.Aspect())
	forward := basis.Pole.Scale(-1) // camera looks along -pole
	denom := dir.Dot(forward)
	if denom == 0 {
		return wire.Vec3{}, false
	}
	t := ui.SceneSphere.Center.Sub(eye).Dot(forward) / denom
	hit := eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return wire.Vec3{}, false
	}
	return hit, true
}

// SetHoverUI sets the Go-owned hover state and MESSAGES the affected node(s) to update their
// OWN hovered bit — no shared/republished map. sendMove is threaded through as a bound func
// value (mirrors DropPointFromNDC's callers).
func (ui *UIState) SetHoverUI(sendMove func(id string, msg movemsg.Msg), node, port string, isInput bool) {
	prevNode := ui.Sel.HoverNode
	ui.Sel.HoverNode, ui.Sel.HoverPort, ui.Sel.HoverInput = node, port, isInput
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, movemsg.Msg{Kind: movemsg.KindHover, NodeID: prevNode, Bool: false})
	}
	if node != "" {
		sendMove(node, movemsg.Msg{Kind: movemsg.KindHover, NodeID: node, Bool: true, Port: port, IsInput: isInput})
	}
}

// DragPlaneHit unprojects ev's pointer onto the camera-facing plane through
// ui.Gest.DragStartCenter, returning the world-space hit. Shared by Wiring's
// commitDragStart (which uses it ONCE to capture DragGrabOffset) and applyNodeDragTarget
// (which uses it every move) so both project against the exact same plane. Returns ok=false
// when the ray is parallel to the plane or the hit is non-finite.
func (ui *UIState) DragPlaneHit(ev inputcodec.RawInputMsg) (hit wire.Vec3, ok bool) {
	g := &ui.Gest
	vp := ui.VP.Viewpoint
	eye := geom.EyeOf(vp)
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	nx, ny := g.PixelToNDC(ev.X, ev.Y)
	dir := geom.RayDirThroughNDC(nx, ny, basis, ev.Fov, g.Rect.Aspect())
	forward := basis.Pole.Scale(-1) // camera looks along -pole
	denom := dir.Dot(forward)
	if denom == 0 {
		return wire.Vec3{}, false
	}
	t := g.DragStartCenter.Sub(eye).Dot(forward) / denom
	hit = eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return wire.Vec3{}, false
	}
	return hit, true
}

// OrbitViewpoint/OrbitLockedViewpoint/ZoomViewpoint promote onto the owned VP
// (gesturefsm.ViewpointState) — pure moves from Wiring's viewpoint_state.go, unchanged
// except for the receiver's new package.
func (ui *UIState) OrbitViewpoint(from, to geom.Dir, tr *T.Trace) {
	ui.VP.OrbitViewpoint(from, to, tr)
}
func (ui *UIState) OrbitLockedViewpoint(from, to geom.Dir, tr *T.Trace) {
	ui.VP.OrbitLockedViewpoint(from, to, tr)
}
func (ui *UIState) ZoomViewpoint(factor float64, tr *T.Trace) {
	ui.VP.ZoomViewpoint(factor, tr)
}

// RefuseStructuralEdit reports a refused create/delete (Wiring's scene_structure.go
// CreateNode/DeleteNode). It goes to STDERR, which the extension host pipes to the sim's
// output channel and .probe/go-errors.jsonl. Nothing is written and the run does not end.
// It mutates EditRefused only; every call site follows it with an EmitViewFrame(nil) of its
// own — the VIEW frame is emitted by the caller, per
// docs/planning/movedispatch-decomposition.md's write-then-emit split.
func (ui *UIState) RefuseStructuralEdit(why string) {
	fmt.Fprintf(os.Stderr, "structural edit refused: %s\n", why)
	// …and SAY SO ON SCREEN. The reason belongs in the log; that the edit was refused at all
	// is the part a person cannot otherwise see, since the scene looks exactly as it did.
	ui.EditRefused++
}
