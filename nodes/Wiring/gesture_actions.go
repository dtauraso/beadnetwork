package Wiring

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
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
func (md *MoveDispatch) beginSphereRotation(ev inputcodec.RawInputMsg) {
	g := &md.ui.gest
	vp := md.ui.vp.Viewpoint
	pivot := geom.FocusAhead(vp, md.lq.heldCenters(md))
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
	_, csRadius := geom.ContentSphereOf(md.lq.heldCenters(md))
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
		if n, ok := md.nodeFromHit(ev.Hit); ok {
			node = n
		}
	}
	md.setHover(node, "", false, tr)
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
	md.OrbitViewpoint(geom.WorldDirToAngles(currDir), geom.WorldDirToAngles(prevDir), tr)
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
	md.OrbitLockedViewpoint(geom.WorldDirToAngles(currDir), geom.WorldDirToAngles(prevDir), tr)
}

// dragPlaneHit unprojects ev's pointer onto the camera-facing plane through
// g.dragStartCenter, returning the world-space hit. Shared by commitDragStart (which uses
// it ONCE to capture g.dragGrabOffset) and applyNodeDragTarget (which uses it every move) so
// both project against the exact same plane instead of two copies that can drift apart.
// Returns ok=false when the ray is parallel to the plane or the hit is non-finite.
func (md *MoveDispatch) dragPlaneHit(ev inputcodec.RawInputMsg) (hit vec3, ok bool) {
	g := &md.ui.gest
	vp := md.ui.vp.Viewpoint
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
func (md *MoveDispatch) applyNodeDragTarget(ev inputcodec.RawInputMsg) bool {
	g := &md.ui.gest
	hit, ok := md.dragPlaneHit(ev)
	if !ok {
		return false
	}
	md.lq.RootMove(md, g.dragNode, hit.Add(g.dragGrabOffset))
	return true
}

// setHover is the shared dedupe+emit hover write; updateHover (pointer path) is its
// one caller.
func (md *MoveDispatch) setHover(node, port string, isInput bool, tr *T.Trace) {
	if node == md.ui.sel.HoverNode && port == md.ui.sel.HoverPort && isInput == md.ui.sel.HoverInput {
		return // no change → no re-emit (dedupe)
	}
	// ui_state.go's setHoverUI is the AUTHORITATIVE write: it sets md.ui.sel's hover
	// fields (mutated only by this goroutine) and MESSAGES the affected
	// node(s) to set their OWN hovered bit — no shared/republished map.
	md.ui.setHoverUI(md.sendMove, node, port, isInput)
	// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): this same goroutine also writes
	// its own VIEW frame directly, carrying this one hover event resolved to buffer rows
	// (mirrors owner_events.go's pattern for every other per-owner stream).
	nodeRow := int32(-1)
	if r, ok := md.RT.NodeRowFor(node); ok {
		nodeRow = r
	}
	// portRow is always -1: a port has no buffer row of its own any more
	// (docs/bead-model/channels-not-ports.md — hover addresses the node, not a port).
	portRow := int32(-1)
	value := int32(0)
	if isInput {
		value = 1
	}
	md.emitViewFrame([]wire.RowEvent{{Kind: T.KindHover, NodeRow: nodeRow, PortRow: portRow, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: value}})
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
		md.setSelectionUI("", "")
		md.emitSelectViewFrame("")
		return
	}
	if ev.Hit.Kind == "edge" {
		if label, ok := md.edgeFromHit(ev.Hit); ok {
			md.setSelectionUI("", label)
			// An edge selection carries no NodeRow (see decodeEventLine's "select" case,
			// buffer-log.ts — it never reads EdgeRow for this kind), mirroring the
			// KindSelect{Edge: label, Node: ""} shape exactly.
			md.emitSelectViewFrame("")
			return
		}
		// Unresolvable edge hit → clear selection rather than leaving stale state.
	}

	var node string
	if ev.Hit.Kind == "node" {
		if n, ok := md.nodeFromHit(ev.Hit); ok {
			node = n
		}
	}
	md.setSelectionUI(node, "")
	md.emitSelectViewFrame(node)
}

// emitSelectViewFrame is applySelect's decentralized-view-frame counterpart (Step C,
// memory/feedback_no_single_writer_bridge.md): writes this goroutine's own VIEW frame carrying the one
// select event just logged via tr.Select/tr.SelectEdge above.
func (md *MoveDispatch) emitSelectViewFrame(node string) {
	nodeRow := int32(-1)
	if node != "" {
		if r, ok := md.RT.NodeRowFor(node); ok {
			nodeRow = r
		}
	}
	md.emitViewFrame([]wire.RowEvent{{Kind: T.KindSelect, NodeRow: nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}})
}
