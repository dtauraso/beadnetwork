package Wiring

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
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
func beginSphereRotation(ui *uiState, heldCenters func() map[string]vec3, ev inputcodec.RawInputMsg) {
	g := &ui.gest
	vp := ui.vp.Viewpoint
	pivot := geom.FocusAhead(vp, heldCenters())
	g.rotPivot = pivot

	eye := geom.EyeOf(vp)
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	ndcX, ndcY, _ := geom.ProjectNDC(pivot, eye, basis, ev.Fov, g.rect.aspect())
	g.rotCx = ((ndcX+1)/2)*g.rect.width + g.rect.left
	g.rotCy = ((-ndcY+1)/2)*g.rect.height + g.rect.top

	// Rotate sensitivity is ANCHORED TO THE ON-SCREEN CONTENT-SPHERE RADIUS: pixels-per-radian
	// scales by csRadius/pivotDist (the sphere's angular size), so a quarter-turn (pi/2) is
	// reached by dragging one on-screen content-sphere radius, at every zoom level. Without the
	// anchor, pi/2 required dragging nearly the full screen height and felt unreachable.
	_, csRadius := geom.ContentSphereOf(heldCenters())
	pivotDist := eye.Sub(pivot).Length()
	fovRad := ev.Fov * math.Pi / 180
	rpx := (g.rect.height / 2) / math.Tan(fovRad/2)
	if pivotDist > 0 {
		rpx *= csRadius / pivotDist
	}
	g.rotPxPerRad = math.Max(rpx*(2/math.Pi), 1)
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
	if events, changed := setHover(&md.ui, sendMoveFn, &md.RT, node, "", false, tr); changed {
		md.emitViewFrame(events)
	}
}

// seedOrbitPivot installs the frozen pivot as the viewpoint pivot (mirrors the TS
// sendViewpointSet at rotation start): pos/up/r recompute about the new pivot so the
// subsequent orbit is rigid about it.
func (md *MoveDispatch) seedOrbitPivot(pivot vec3) {
	vp := md.ui.vp.Viewpoint
	eye := geom.EyeOf(vp)
	r := eye.Sub(pivot).Length()
	pos := geom.WorldDirToAngles(eye.Sub(pivot))
	md.SetViewpoint(pivot, r, pos, vp.Up)
}

// applyOrbit mirrors the "rotating" branch of interaction-handlers.ts handlePointerMove:
// map prev/curr cursor pixels through the frozen sphere frame to world directions and orbit
// (curr → prev), so the grabbed direction follows the cursor.
func (md *MoveDispatch) applyOrbit(ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &md.ui.gest
	vp := md.ui.vp.Viewpoint
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := geom.ScreenToPolar(g.prevX-g.rotCx, g.prevY-g.rotCy, g.rotPxPerRad)
	curr := geom.ScreenToPolar(ev.X-g.rotCx, ev.Y-g.rotCy, g.rotPxPerRad)
	prevDir := geom.ToWorldDir(basis, prev)
	currDir := geom.ToWorldDir(basis, curr)
	md.ui.OrbitViewpoint(geom.WorldDirToAngles(currDir), geom.WorldDirToAngles(prevDir), tr)
	md.emitViewFrame(cameraViewEvent())
}

// applyOrbitLocked mirrors the "handhold-rotating" branch of interaction-handlers.ts
// handlePointerMove: identical prev/curr → world-direction mapping as applyOrbit, but the
// arc is applied through OrbitLockedViewpoint, which locks the rotation axis on the first
// call and reuses it (constrained "disk" orbit). The lock clears on the next SetViewpoint.
func (md *MoveDispatch) applyOrbitLocked(ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &md.ui.gest
	vp := md.ui.vp.Viewpoint
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := geom.ScreenToPolar(g.prevX-g.rotCx, g.prevY-g.rotCy, g.rotPxPerRad)
	curr := geom.ScreenToPolar(ev.X-g.rotCx, ev.Y-g.rotCy, g.rotPxPerRad)
	prevDir := geom.ToWorldDir(basis, prev)
	currDir := geom.ToWorldDir(basis, curr)
	md.ui.OrbitLockedViewpoint(geom.WorldDirToAngles(currDir), geom.WorldDirToAngles(prevDir), tr)
	md.emitViewFrame(cameraViewEvent())
}

// dragPlaneHit unprojects ev's pointer onto the camera-facing plane through
// g.dragStartCenter, returning the world-space hit. Shared by commitDragStart (which uses
// it ONCE to capture g.dragGrabOffset) and applyNodeDragTarget (which uses it every move) so
// both project against the exact same plane instead of two copies that can drift apart.
// Returns ok=false when the ray is parallel to the plane or the hit is non-finite.
func (ui *uiState) dragPlaneHit(ev inputcodec.RawInputMsg) (hit vec3, ok bool) {
	g := &ui.gest
	vp := ui.vp.Viewpoint
	eye := geom.EyeOf(vp)
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	nx, ny := g.pixelToNDC(ev.X, ev.Y)
	dir := geom.RayDirThroughNDC(nx, ny, basis, ev.Fov, g.rect.aspect())
	forward := basis.Pole.Scale(-1) // camera looks along -pole
	denom := dir.Dot(forward)
	if denom == 0 {
		return vec3{}, false
	}
	t := g.dragStartCenter.Sub(eye).Dot(forward) / denom
	hit = eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return vec3{}, false
	}
	return hit, true
}

// applyNodeDragTarget mirrors the "dragging" branch: unproject the pointer onto a
// camera-facing plane through the node's start center, giving a free world target, then
// RootMove the node to that target PLUS the grab offset captured once at drag start (Go
// snaps it to the parent sphere). Adding the offset here — instead of moving the node's
// center straight to the hit — is what keeps the point you grabbed under the cursor instead
// of the node teleporting so its center lands there. Returns false if the ray is parallel
// to the plane.
func applyNodeDragTarget(ui *uiState, rootMove func(id string, target vec3) bool, ev inputcodec.RawInputMsg) bool {
	g := &ui.gest
	hit, ok := ui.dragPlaneHit(ev)
	if !ok {
		return false
	}
	rootMove(g.dragNode, hit.Add(g.dragGrabOffset))
	return true
}

// setHover is the shared dedupe+mutate hover write; updateHover (pointer path) is its one
// caller. It mutates md.ui's hover fields and reports whether they changed plus the one
// hover RowEvent to emit; the caller (the view-owner goroutine) emits it — this method
// itself never calls emitViewFrame, per docs/planning/movedispatch-decomposition.md's
// write-then-emit split.
func setHover(ui *uiState, sendMoveFn func(id string, msg movemsg.Msg), RT *rowtables.RowTables, node, port string, isInput bool, tr *T.Trace) (events []wire.RowEvent, changed bool) {
	if node == ui.sel.HoverNode && port == ui.sel.HoverPort && isInput == ui.sel.HoverInput {
		return nil, false // no change → no re-emit (dedupe)
	}
	// ui_state.go's setHoverUI is the AUTHORITATIVE write: it sets ui.sel's hover
	// fields (mutated only by this goroutine) and MESSAGES the affected
	// node(s) to set their OWN hovered bit — no shared/republished map.
	ui.setHoverUI(sendMoveFn, node, port, isInput)
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
		setSelectionUI(&md.ui, &md.mr, md.ctx, "", "")
		md.emitViewFrame(md.RT.SelectViewEvent(""))
		return
	}
	if ev.Hit.Kind == "edge" {
		if label, ok := md.RT.EdgeFromHit(ev.Hit); ok {
			setSelectionUI(&md.ui, &md.mr, md.ctx, "", label)
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
	setSelectionUI(&md.ui, &md.mr, md.ctx, node, "")
	md.emitViewFrame(md.RT.SelectViewEvent(node))
}
