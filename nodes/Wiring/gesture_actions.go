package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"math"

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
func (md *MoveDispatch) beginSphereRotation(ev rawInputMsg) {
	g := &md.ui.gest
	vp := md.ui.vp.viewpoint
	pivot := focusAhead(vp, md.heldCenters())
	g.rotPivot = pivot

	eye := eyeOf(vp)
	basis := basisFromViewpoint(vp.pos, vp.up)
	ndcX, ndcY, _ := projectNDC(pivot, eye, basis, ev.Fov, g.rect.aspect())
	g.rotCx = ((ndcX+1)/2)*g.rect.width + g.rect.left
	g.rotCy = ((-ndcY+1)/2)*g.rect.height + g.rect.top

	// Rotate sensitivity is ANCHORED TO THE ON-SCREEN CONTENT-SPHERE RADIUS: pixels-per-radian
	// scales by csRadius/pivotDist (the sphere's angular size), so a quarter-turn (pi/2) is
	// reached by dragging one on-screen content-sphere radius, at every zoom level. Without the
	// anchor, pi/2 required dragging nearly the full screen height and felt unreachable.
	_, csRadius := contentSphereOf(md.heldCenters())
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
// the node's / port's Hovered column. Hover is node+port only (edges do not hover on the
// pre-branch path). Deduping on the (node, port, isInput) triple keeps a still pointer and a
// same-entity drag from re-emitting a snapshot each pointer-move (no new flood — Go already
// emits per raw-input; a hover only fires on a genuine target change). An empty / edge / other
// hit clears hover.
func (md *MoveDispatch) updateHover(ev rawInputMsg, tr *T.Trace) {
	var node, port string
	var isInput bool
	switch ev.Hit.Kind {
	case "port":
		if n, p, in, ok := md.portFromHit(ev.Hit); ok {
			node, port, isInput = n, p, in
		}
	case "torus":
		// The concentric hover ring emphasizes the TORUS handle, so it lights only when the
		// cursor is actually on the ring — NOT on the node body. A plain "node"-body hit
		// deliberately falls through here and clears hover (node-body hover feedback is a
		// separate concern, not wired yet).
		if n, ok := md.nodeFromHit(ev.Hit); ok {
			node = n
		}
	}
	md.setHover(node, port, isInput, tr)
}

// seedOrbitPivot installs the frozen pivot as the viewpoint pivot (mirrors the TS
// sendViewpointSet at rotation start): pos/up/r recompute about the new pivot so the
// subsequent orbit is rigid about it.
func (md *MoveDispatch) seedOrbitPivot(pivot vec3) {
	vp := md.ui.vp.viewpoint
	eye := eyeOf(vp)
	r := eye.Sub(pivot).Length()
	pos := worldDirToAngles(eye.Sub(pivot))
	md.SetViewpoint(pivot, r, pos, vp.up)
}

// applyOrbit mirrors the "rotating" branch of interaction-handlers.ts handlePointerMove:
// map prev/curr cursor pixels through the frozen sphere frame to world directions and orbit
// (curr → prev), so the grabbed direction follows the cursor.
func (md *MoveDispatch) applyOrbit(ev rawInputMsg, tr *T.Trace) {
	g := &md.ui.gest
	vp := md.ui.vp.viewpoint
	basis := basisFromViewpoint(vp.pos, vp.up)
	prev := screenToPolar(g.prevX-g.rotCx, g.prevY-g.rotCy, g.rotPxPerRad)
	curr := screenToPolar(ev.X-g.rotCx, ev.Y-g.rotCy, g.rotPxPerRad)
	prevDir := toWorldDir(basis, prev)
	currDir := toWorldDir(basis, curr)
	md.OrbitViewpoint(worldDirToAngles(currDir), worldDirToAngles(prevDir), tr)
}

// applyOrbitLocked mirrors the "handhold-rotating" branch of interaction-handlers.ts
// handlePointerMove: identical prev/curr → world-direction mapping as applyOrbit, but the
// arc is applied through OrbitLockedViewpoint, which locks the rotation axis on the first
// call and reuses it (constrained "disk" orbit). The lock clears on the next SetViewpoint.
func (md *MoveDispatch) applyOrbitLocked(ev rawInputMsg, tr *T.Trace) {
	g := &md.ui.gest
	vp := md.ui.vp.viewpoint
	basis := basisFromViewpoint(vp.pos, vp.up)
	prev := screenToPolar(g.prevX-g.rotCx, g.prevY-g.rotCy, g.rotPxPerRad)
	curr := screenToPolar(ev.X-g.rotCx, ev.Y-g.rotCy, g.rotPxPerRad)
	prevDir := toWorldDir(basis, prev)
	currDir := toWorldDir(basis, curr)
	md.OrbitLockedViewpoint(worldDirToAngles(currDir), worldDirToAngles(prevDir), tr)
}

// applyPortMove mirrors the "port-move" branch of interaction-handlers.ts handlePointerMove:
// project the pointer ray onto the horizontal plane (normal +z) at the node's ring height
// (z = center.z), take the in-plane direction from center to the hit (z zeroed, matching
// pointerRingAnchor), and apply it as a ring-anchor update via the existing anchor path.
func (md *MoveDispatch) applyPortMove(ev rawInputMsg) {
	g := &md.ui.gest
	hit, ok := md.pointerOnRingPlane(ev, g.portMoveCenter.Z)
	if !ok {
		return
	}
	dx := hit.X - g.portMoveCenter.X
	dy := hit.Y - g.portMoveCenter.Y
	if dx == 0 && dy == 0 {
		return
	}
	md.applyRingAnchor(g.portMoveNode, g.portMovePort, g.portMoveInput, vec3{X: dx, Y: dy, Z: 0})
}

// pointerOnRingPlane intersects the pointer ray with the horizontal plane (normal +z) at
// world height planeZ, mirroring interaction-handlers.ts unprojectToPlane. Returns
// (hit, false) if the ray is parallel to the plane or the result is non-finite.
func (md *MoveDispatch) pointerOnRingPlane(ev rawInputMsg, planeZ float64) (vec3, bool) {
	g := &md.ui.gest
	vp := md.ui.vp.viewpoint
	eye := eyeOf(vp)
	basis := basisFromViewpoint(vp.pos, vp.up)
	nx, ny := g.pixelToNDC(ev.X, ev.Y)
	dir := rayDirThroughNDC(nx, ny, basis, ev.Fov, g.rect.aspect())
	if dir.Z == 0 {
		return vec3{}, false
	}
	t := (planeZ - eye.Z) / dir.Z
	hit := eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return vec3{}, false
	}
	return hit, true
}

// applyNodeDragTarget mirrors the "dragging" branch: unproject the pointer onto a
// camera-facing plane through the node's start center, giving a free world target, then
// RootMove the node (Go snaps it to the parent sphere). Returns false if the ray is parallel
// to the plane.
func (md *MoveDispatch) applyNodeDragTarget(ev rawInputMsg) bool {
	g := &md.ui.gest
	vp := md.ui.vp.viewpoint
	eye := eyeOf(vp)
	basis := basisFromViewpoint(vp.pos, vp.up)
	nx, ny := g.pixelToNDC(ev.X, ev.Y)
	dir := rayDirThroughNDC(nx, ny, basis, ev.Fov, g.rect.aspect())
	forward := basis.pole.Scale(-1) // camera looks along -pole
	denom := dir.Dot(forward)
	if denom == 0 {
		return false
	}
	t := g.dragStartCenter.Sub(eye).Dot(forward) / denom
	hit := eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return false
	}
	md.RootMove(g.dragNode, hit)
	return true
}

// setHover is the shared dedupe+emit hover write; updateHover (pointer path) is its
// one caller.
func (md *MoveDispatch) setHover(node, port string, isInput bool, tr *T.Trace) {
	if node == md.ui.sel.hoverNode && port == md.ui.sel.hoverPort && isInput == md.ui.sel.hoverInput {
		return // no change → no re-emit (dedupe)
	}
	// setHoverUI (node_move.go) is the AUTHORITATIVE write: it sets md.ui.sel's hover
	// fields (mutated only by this goroutine) and MESSAGES the affected
	// node(s) to set their OWN hovered bit — no shared/republished map.
	md.setHoverUI(node, port, isInput)
	// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): this same goroutine also writes
	// its own VIEW frame directly, carrying this one hover event resolved to buffer rows
	// (mirrors owner_events.go's pattern for every other per-owner stream).
	nodeRow := int32(-1)
	if r, ok := md.NodeRowFor(node); ok {
		nodeRow = r
	}
	portRow := int32(-1)
	if port != "" {
		if r, ok := md.PortRowFor(node, port, isInput); ok {
			portRow = r
		}
	}
	value := int32(0)
	if isInput {
		value = 1
	}
	md.emitViewFrame([]wire.RowEvent{{Kind: T.KindHover, NodeRow: nodeRow, PortRow: portRow, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: value}})
}

// applySelect sets the Go-owned selection from a click hit and emits it. Selection is
// single + EXCLUSIVE across nodes and edges: an EDGE hit selects that edge (clearing any
// node selection); a node/port hit selects that node (clearing any edge selection); an
// empty-space hit CLEARS the transient highlight (md.ui.sel.selected / md.ui.sel.selectedEdge) — this is
// the original click-empty-clears behavior.
func (md *MoveDispatch) applySelect(ev rawInputMsg, tr *T.Trace) {
	// setSelectionUI (node_move.go) is the AUTHORITATIVE write, same reasoning as
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
	switch ev.Hit.Kind {
	case "node":
		if n, ok := md.nodeFromHit(ev.Hit); ok {
			node = n
		}
	case "port":
		if n, _, _, ok := md.portFromHit(ev.Hit); ok {
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
		if r, ok := md.NodeRowFor(node); ok {
			nodeRow = r
		}
	}
	md.emitViewFrame([]wire.RowEvent{{Kind: T.KindSelect, NodeRow: nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}})
}

// applyRingAnchor snaps a world-space direction (node center → pointer) to the node's
// nearest ring-anchor index and mail-sorts a moveMsgKindAnchor to the node's mover AND
// every incident edge mover — the SAME dispatch the op=update kind=node attr=anchor path
// uses (applyUpdate). Disk persistence is NOT done here: the node's own mover persists the
// snapped index to its own port file, on its own goroutine, when it processes the
// moveMsgKindAnchor sent below (node_mover.go handle → persistPortAnchor) — this function
// only routes the message.
//
// This sends directly into the targets' dedicated extIn channels, bypassing the
// enqueueFor/pending-retry split every mover's OWN handler goroutine must use for its
// sends. That split exists to
// prevent two mutually-adjacent MOVER goroutines from deadlocking each other — both
// mid-handle, each blocked sending into the other's full channel, while neither is
// draining its own (draining only resumes after handle returns). applyRingAnchor runs on
// the stdin/gesture goroutine, not on any mover's own handler: it is never itself the
// target of one of these sends, so it cannot be a link in that cycle — a block here can
// only ever be "wait for the target's own run loop to read", never "wait for a goroutine
// that is itself waiting on us". That is a real, structural reason this exemption holds,
// not just "it hasn't happened yet".
func (md *MoveDispatch) applyRingAnchor(node, port string, isInput bool, dir vec3) {
	anchorID := snapToRingAnchorIndex(md.NodeKind(node), dir)
	msg := moveMsg{Kind: moveMsgKindAnchor, NodeID: node, Port: port, IsInput: isInput, AnchorId: anchorID}
	if nm, ok := md.mr.nodeMovers[node]; ok {
		nm.extIn <- msg
	}
	for _, em := range md.mr.edgeMovers {
		incident := (isInput && em.dstID == node && em.dstH == port) ||
			(!isInput && em.srcID == node && em.srcH == port)
		if !incident {
			continue
		}
		em.extIn <- msg
	}
	// The snapped anchor index is persisted by the node's OWN mover, on its own
	// goroutine, as it processes the moveMsgKindAnchor sent above (node_mover.go
	// handle's moveMsgKindAnchor case → persistPortAnchor) — not reached into from
	// here (docs/planning/decentralized-persistence.md "The model").
}
