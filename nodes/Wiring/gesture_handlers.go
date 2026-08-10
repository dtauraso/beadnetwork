package Wiring

import (
	"math"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

// gesture_handlers.go — the per-event PHASE HANDLERS the dispatch table in gesture_dispatch.go
// calls into: gestHome, gestPointerDown, gestPointerMove, gestPointerUp, gestWheel. These own
// the FSM's transitions; the state they mutate is declared in gesture.go. Pointer-down's hit
// classification table lives in gesture_hitclassify.go.

// gestHome handles a "home" (fit-to-content) command: Go frames ALL nodes from its OWN held
// geometry with the SAME fit math the TS HomeButton used (homeFitPose), then installs the
// result via SetViewpoint + EmitViewpoint — the exact path a gesture uses. The FSM's own
// viewpoint IS the framed pose; EmitViewpoint streams it out on this goroutine's own
// per-owner VIEW frame (the buffer VIEW stream) and it persists on the polar save path. TS
// sent no pose, only render context (fov + aspect). Because the FSM's own viewpoint now IS
// the framed pose, the next orbit/pan/zoom builds on it (no snap-back). Does nothing when
// there are no nodes, mirroring HomeButton's early return.
func (md *MoveDispatch) gestHome(ev inputcodec.RawInputMsg, tr *T.Trace) {
	centers := md.heldCenters()
	radius := make(map[string]float64, len(centers))
	for id := range centers {
		radius[id] = md.nodeBodyRadius(id)
	}
	pivot, r, pos, up, ok := geom.HomeFitPose(centers, radius, ev.Fov, md.ui.gest.rect.aspect())
	if !ok {
		return
	}
	md.SetViewpoint(pivot, r, pos, up)
	md.EmitViewpoint(tr)
}

func (md *MoveDispatch) gestPointerDown(ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &md.ui.gest
	g.downX, g.downY = ev.X, ev.Y
	g.prevX, g.prevY = ev.X, ev.Y
	g.button = ev.Button
	g.secondary = ev.Button == 2 // two-finger trackpad tap → always a tap-select
	g.phase = gestPending
	g.emptyDown = false
	g.dragNode = ""
	g.handholdDown = false

	if h, ok := hitClassifiers[ev.Hit.Kind]; ok {
		h(md, g, ev)
	}
}

func (md *MoveDispatch) gestPointerMove(ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &md.ui.gest
	if g.phase == gestIdle {
		return
	}
	dx := ev.X - g.downX
	dy := ev.Y - g.downY
	dist := math.Hypot(dx, dy)

	// Click vs. drag is discriminated by MOVEMENT ITSELF, not a distance threshold: any
	// actual displacement from the press point commits (dist > 0, not merely "a move event
	// arrived" — some input stacks emit a move AT the press coordinates, which must not
	// commit). A prior version gated this on dist > a slop constant (gestureMoveSlopPx,
	// deleted), which doubled as a click-vs-drag discriminator AND a hidden fix for a
	// different defect: before the grab-offset fix (commitDragStart's dragGrabOffset), the
	// node's center would JUMP to the cursor the instant the slop was crossed, so a large
	// threshold delayed that jump rather than removing it. With the grab offset preserved,
	// engaging on the very first pixel of movement is smooth, so the threshold now only
	// costs responsiveness — it can go. Trade-off accepted deliberately: hand tremor during
	// a click now registers as a sub-pixel drag, which is harmless because node positions
	// quantize to the bead lattice, so sub-lattice movement usually resolves to no change.
	//
	// A secondary (two-finger) press never becomes a drag/rotate — it is a tap-select, so
	// it stays gestPending through any finger drift and resolves on pointer-up.
	if g.phase == gestPending && dist > 0 && !g.secondary {
		for _, edge := range commitEdges {
			if edge.guard(g) {
				edge.action(md, g, ev, tr)
				g.phase = edge.to
				break
			}
		}
	}

	if apply, ok := applyAction[g.phase]; ok {
		apply(md, g, ev, tr)
	}
}

func (md *MoveDispatch) gestPointerUp(ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	g := &md.ui.gest
	switch {
	case g.phase == gestDragging:
		md.applyNodeDragTarget(ev) // final target flush
	case g.phase == gestHandhold, g.phase == gestRotating:
		// Rotation completed (free or handhold-constrained): nothing to flush.
	case g.phase == gestPending:
		// Click → Go-owned selection. A node hit selects it; empty space clears the
		// selection. md.ui.sel.selected is the authoritative selection; Select() emits it so the
		// buffer snapshot marks the node's Selected column.
		md.applySelect(ev, tr)
	}
	wasDragging := g.phase == gestDragging
	// Capture BEFORE reset() clears it (below) — movemsg.KindDragEnd must name the node
	// that was actually dragged, and reset() zeroes g.dragNode unconditionally.
	draggedNode := g.dragNode
	g.reset(&md.ui.vp.Viewpoint)
	if wasDragging {
		// The drag just ended: g.dragNode is now "" (cleared by reset above), so the
		// Overlay block's DragNodeRow column must go back to -1 promptly rather than
		// waiting for the next unrelated view-frame emit. Mirrors commitDragStart's own
		// emitViewFrame call at drag START.
		md.emitViewFrame(nil)
		// "done dragging" (PLAN.md) — mirrors commitDragStart's own movemsg.KindDragStart
		// send. Sent on EVERY path a drag ends by (this is the FSM's one drag-end exit),
		// so a chain bead this node woke can never be left on machine time — see
		// movemsg.KindDragEnd's own doc comment.
		if draggedNode != "" {
			md.sendMove(draggedNode, movemsg.Msg{Kind: movemsg.KindDragEnd, NodeID: draggedNode})
		}
	}
}

// gestWheel mirrors interaction-handlers.ts handleWheelNative: ctrl+wheel = zoom-to-cursor
// dolly (expressed as a PAN in the polar model — a pivot translation, not a radius change),
// plain wheel = screen-space pan. Both first seed the viewpoint to region-focus, then pan.
func (md *MoveDispatch) gestWheel(ev inputcodec.RawInputMsg, tr *T.Trace) {
	vp := md.ui.vp.Viewpoint
	eye := geom.EyeOf(vp)
	pivot := geom.RegionFocus(vp, md.heldCenters())

	if ev.Ctrl {
		// Zoom-to-cursor: move the camera TOWARD the node under the cursor along the cursor→node
		// line, KEEPING the look direction — so that node stays fixed under the mouse. It does NOT
		// re-aim: re-aiming (snapping the camera to look straight at the node) is what recentered
		// the view and threw the cursor off. PanViewpoint translates the whole camera (pivot+eye
		// ride together); pos/up are unchanged, so the node keeps projecting to the same pixel.
		// The cursor→node pick is a screen-space selection at the input boundary (projectNDC).
		mouseNdcX, mouseNdcY := md.ui.gest.pixelToNDC(ev.X, ev.Y)
		basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
		aspect := md.ui.gest.rect.aspect()
		target := pivot
		best := math.Inf(1)
		for _, c := range md.heldCenters() {
			nx, ny, inFront := geom.ProjectNDC(c, eye, basis, ev.Fov, aspect)
			if !inFront {
				continue
			}
			if d := math.Hypot(nx-mouseNdcX, ny-mouseNdcY); d < best {
				best = d
				target = c
			}
		}
		toTarget := target.Sub(eye)
		distP := toTarget.Length()
		rayDir := geom.AnglesToWorldOffset(1, vp.Pos.Theta, vp.Pos.Phi).Scale(-1) // forward, if AT the node
		if distP > 1e-9 {
			rayDir = toTarget.Scale(1 / distP)
		}
		// Move the eye ALONG the cursor→node ray. amt>0 = toward the node (zoom in). The step is a
		// fraction of the remaining distance (fast approach when far), FLOORED at a scene-scaled
		// minimum so you can push THROUGH the node instead of asymptotically creeping to it — a
		// pilot camera flies past nodes. No stop-short clamp.
		amt := 1 - math.Pow(geom.GestureZoomBase, ev.DeltaY)
		step := distP * amt
		if minStep := vp.R * (geom.GestureZoomBase - 1); math.Abs(step) < minStep {
			step = math.Copysign(minStep, amt)
		}
		md.PanViewpoint(rayDir.Scale(step), tr)
		return
	}

	// Plain wheel = LATERAL pan = STRAFE THE CAMERA (free-camera model): the camera body slides
	// sideways through the fixed scene. Pan SPEED is scaled by the camera's OWN focal distance
	// (vp.r), NOT by eye-to-nearest-content — the latter collapses when zoom dollies the eye up
	// to a node, which is exactly what made pan crawl after zooming in (and coupled pan to zoom).
	// vp.r is a stable scene-scale property (set by home/framing, unchanged by the dolly), so pan
	// stays a usable pilot speed at any zoom. The displacement is built in polar; PanViewpoint
	// translates pivot+eye together with the look direction unchanged. The scene does not move.
	fovRad := ev.Fov * math.Pi / 180
	worldPerPixel := (2 * vp.R * math.Tan(fovRad/2)) / md.ui.gest.rect.height
	disp := geom.PanDisplacementPolar(vp.Pos, vp.Up, ev.DeltaX, ev.DeltaY, worldPerPixel)
	md.PanViewpoint(disp, tr)
}
