package Wiring

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// gesture_actions.go — the LEAF ACTIONS invoked by the gesture FSM's phase handlers in
// gesture.go: orbit/rotation seeding, drag/port-move application, hover, and select. This
// file owns no FSM state transitions; it only performs the behavior a phase handler decided
// to invoke.

// beginSphereRotation freezes the orbit pivot, its screen-pixel center, and pixels-per-radian
// for the whole gesture. The pivot is the CONTENT DIRECTLY AHEAD (focusAhead): the node the
// camera is most pointed at, at its depth on the view-center ray. So rotate orbits whatever you
// have flown to and centered (fly to a node → rotate spins around it), the orbit depth tracks
// what you look at, and — because the pivot is on the view axis — it does not re-aim the camera.
func beginSphereRotation(ui *viewstate.UIState, heldCenters func() map[string]vec3, ev inputcodec.RawInputMsg) {
	g := &ui.Gest
	vp := ui.VP.Viewpoint
	pivot := geom.FocusAhead(vp, heldCenters())
	g.RotPivot = pivot

	eye := geom.EyeOf(vp)
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	ndcX, ndcY, _ := geom.ProjectNDC(pivot, eye, basis, ev.Fov, g.Rect.Aspect())
	g.RotCx = ((ndcX+1)/2)*g.Rect.Width + g.Rect.Left
	g.RotCy = ((-ndcY+1)/2)*g.Rect.Height + g.Rect.Top

	// Rotate sensitivity is ANCHORED TO THE ON-SCREEN CONTENT-SPHERE RADIUS: pixels-per-radian
	// scales by csRadius/pivotDist (the sphere's angular size), so a quarter-turn (pi/2) is
	// reached by dragging one on-screen content-sphere radius, at every zoom level. Without the
	// anchor, pi/2 required dragging nearly the full screen height and felt unreachable.
	_, csRadius := geom.ContentSphereOf(heldCenters())
	pivotDist := eye.Sub(pivot).Length()
	fovRad := ev.Fov * math.Pi / 180
	rpx := (g.Rect.Height / 2) / math.Tan(fovRad/2)
	if pivotDist > 0 {
		rpx *= csRadius / pivotDist
	}
	g.RotPxPerRad = math.Max(rpx*(2/math.Pi), 1)
}

// updateHover resolves the entity under the pointer from the raycast hit and, WHEN IT
// CHANGES, records it as the Go-owned hover and emits KindHover so the buffer snapshot marks
// the node's Hovered column. Hover is node-only now — a port is a load-time channel-binding
// ROLE (docs/bead-model/channels-not-ports.md), never drawn or raycast-hit, so the old "port" hit branch
// is gone. Deduping on the node keeps a still pointer and a same-entity drag from re-emitting
// a snapshot each pointer-move (no new flood — Go already emits per raw-input; a hover only
// fires on a genuine target change). An empty / edge / other hit clears hover.
func (md *MoveDispatch) updateHover(ev inputcodec.RawInputMsg, tr *T.Trace) {
	var node string
	switch ev.Hit.Kind {
	case "torus":
		// The concentric hover ring emphasizes the TORUS handle, so it lights only when the
		// cursor is actually on the ring — NOT on the node body. A plain "node"-body hit
		// deliberately falls through here and clears hover (node-body hover feedback is a
		// separate concern, not wired yet).
		if n, ok := md.RT.NodeFromHit(ev.Hit); ok {
			node = n
		}
	}
	mr, ctx := &md.mr, md.ctx
	sendMoveFn := func(id string, msg movemsg.Msg) { sendMove(mr, ctx, id, msg) }
	if events, changed := setHover(&md.UI, sendMoveFn, &md.RT, node, "", false, tr); changed {
		md.emitViewFrame(events)
	}
}

// seedOrbitPivot installs the frozen pivot as the viewpoint pivot (mirrors the TS
// sendViewpointSet at rotation start): pos/up/r recompute about the new pivot so the
// subsequent orbit is rigid about it.
func (md *MoveDispatch) seedOrbitPivot(pivot vec3) {
	vp := md.UI.VP.Viewpoint
	eye := geom.EyeOf(vp)
	r := eye.Sub(pivot).Length()
	pos := geom.WorldDirToAngles(eye.Sub(pivot))
	md.UI.VP.SetViewpoint(pivot, r, pos, vp.Up)
}

// applyOrbit mirrors the "rotating" branch of interaction-handlers.ts handlePointerMove:
// map prev/curr cursor pixels through the frozen sphere frame to world directions and orbit
// (curr → prev), so the grabbed direction follows the cursor.
func (md *MoveDispatch) applyOrbit(ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &md.UI.Gest
	vp := md.UI.VP.Viewpoint
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := geom.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := geom.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := geom.ToWorldDir(basis, prev)
	currDir := geom.ToWorldDir(basis, curr)
	md.UI.OrbitViewpoint(geom.WorldDirToAngles(currDir), geom.WorldDirToAngles(prevDir), tr)
	md.emitViewFrame(cameraViewEvent())
}

// applyOrbitLocked mirrors the "handhold-rotating" branch of interaction-handlers.ts
// handlePointerMove: identical prev/curr → world-direction mapping as applyOrbit, but the
// arc is applied through OrbitLockedViewpoint, which locks the rotation axis on the first
// call and reuses it (constrained "disk" orbit). The lock clears on the next SetViewpoint.
func (md *MoveDispatch) applyOrbitLocked(ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &md.UI.Gest
	vp := md.UI.VP.Viewpoint
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := geom.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := geom.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := geom.ToWorldDir(basis, prev)
	currDir := geom.ToWorldDir(basis, curr)
	md.UI.OrbitLockedViewpoint(geom.WorldDirToAngles(currDir), geom.WorldDirToAngles(prevDir), tr)
	md.emitViewFrame(cameraViewEvent())
}

// dragPlaneHit unprojects ev's pointer onto the camera-facing plane through
// g.DragStartCenter, returning the world-space hit. Shared by commitDragStart (which uses
// it ONCE to capture g.DragGrabOffset) and applyNodeDragTarget (which uses it every move) so
// both project against the exact same plane instead of two copies that can drift apart.
// Returns ok=false when the ray is parallel to the plane or the hit is non-finite.
// applyNodeDragTarget mirrors the "dragging" branch: unproject the pointer onto a
// camera-facing plane through the node's start center, giving a free world target, then
// RootMove the node to that target PLUS the grab offset captured once at drag start (Go
// snaps it to the parent sphere). Adding the offset here — instead of moving the node's
// center straight to the hit — is what keeps the point you grabbed under the cursor instead
// of the node teleporting so its center lands there. Returns false if the ray is parallel
// to the plane.
func applyNodeDragTarget(ui *viewstate.UIState, rootMove func(id string, target vec3) bool, ev inputcodec.RawInputMsg) bool {
	g := &ui.Gest
	hit, ok := ui.DragPlaneHit(ev)
	if !ok {
		return false
	}
	rootMove(g.DragNode, hit.Add(g.DragGrabOffset))
	return true
}

// setHover is the shared dedupe+mutate hover write; updateHover (pointer path) is its one
// caller. It mutates md.ui's hover fields and reports whether they changed plus the one
// hover RowEvent to emit; the caller (the view-owner goroutine) emits it — this method
// itself never calls emitViewFrame, per docs/planning/movedispatch-decomposition.md's
// write-then-emit split.
func setHover(ui *viewstate.UIState, sendMoveFn func(id string, msg movemsg.Msg), RT *rowtables.RowTables, node, port string, isInput bool, tr *T.Trace) (events []wire.RowEvent, changed bool) {
	if node == ui.Sel.HoverNode && port == ui.Sel.HoverPort && isInput == ui.Sel.HoverInput {
		return nil, false // no change → no re-emit (dedupe)
	}
	// viewstate's UIState.SetHoverUI is the AUTHORITATIVE write: it sets ui.Sel's hover
	// fields (mutated only by this goroutine) and MESSAGES the affected
	// node(s) to set their OWN hovered bit — no shared/republished map.
	ui.SetHoverUI(sendMoveFn, node, port, isInput)
	nodeRow := int32(-1)
	if r, ok := RT.NodeRowFor(node); ok {
		nodeRow = r
	}
	// portRow is always -1: a port has no buffer row of its own any more
	// (docs/bead-model/channels-not-ports.md — hover addresses the node, not a port).
	portRow := int32(-1)
	value := int32(0)
	if isInput {
		value = 1
	}
	return []wire.RowEvent{{Kind: T.KindHover, NodeRow: nodeRow, PortRow: portRow, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: value}}, true
}

// applySelect sets the Go-owned selection from a click hit and emits it. Selection is
// single + EXCLUSIVE across nodes and edges: an EDGE hit selects that edge (clearing any
// node selection); a node/port hit selects that node (clearing any edge selection); an
// empty-space hit CLEARS the transient highlight (md.ui.sel.Selected / md.ui.sel.SelectedEdge) — this is
// the original click-empty-clears behavior.
func (md *MoveDispatch) applySelect(ev inputcodec.RawInputMsg, tr *T.Trace) {
	// setSelectionUI (move_dispatch_api.go) is the AUTHORITATIVE write, same reasoning as
	// setHoverUI above: it sets md.ui.sel's selection fields (+ latchedNode, mutated only
	// by this goroutine) and MESSAGES the affected node(s)/edge to set their
	// OWN selected/latchedSel bit.
	if ev.Hit.Kind == "empty" {
		setSelectionUI(&md.UI, &md.mr, md.ctx, "", "")
		md.emitViewFrame(md.RT.SelectViewEvent(""))
		return
	}
	if ev.Hit.Kind == "edge" {
		if label, ok := md.RT.EdgeFromHit(ev.Hit); ok {
			setSelectionUI(&md.UI, &md.mr, md.ctx, "", label)
			// An edge selection carries no NodeRow (see decodeEventLine's "select" case,
			// buffer-log.ts — it never reads EdgeRow for this kind), mirroring the
			// KindSelect{Edge: label, Node: ""} shape exactly.
			md.emitViewFrame(md.RT.SelectViewEvent(""))
			return
		}
		// Unresolvable edge hit → clear selection rather than leaving stale state.
	}

	var node string
	if ev.Hit.Kind == "node" {
		if n, ok := md.RT.NodeFromHit(ev.Hit); ok {
			node = n
		}
	}
	setSelectionUI(&md.UI, &md.mr, md.ctx, node, "")
	md.emitViewFrame(md.RT.SelectViewEvent(node))
}
