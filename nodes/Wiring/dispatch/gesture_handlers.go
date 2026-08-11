package dispatch

import (
	"math"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
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
	centers := layoutquant.HeldCenters(md.MR.NodeGeoms(), md.MR.CenterOfNode)
	radius := make(map[string]float64, len(centers))
	for id := range centers {
		radius[id] = md.MR.NodeBodyRadius(id)
	}
	pivot, r, pos, up, ok := geom.HomeFitPose(centers, radius, ev.Fov, md.UI.Gest.Rect.Aspect())
	if !ok {
		return
	}
	md.UI.VP.SetViewpoint(pivot, r, pos, up)
	md.UI.VP.EmitViewpoint(tr)
	md.UI.EmitViewFrame(cameraViewEvent())
}

func (md *MoveDispatch) gestPointerDown(ev inputcodec.RawInputMsg) {
	g := &md.UI.Gest
	g.DownX, g.DownY = ev.X, ev.Y
	g.PrevX, g.PrevY = ev.X, ev.Y
	g.Button = ev.Button
	g.Secondary = ev.Button == 2 // two-finger trackpad tap → always a tap-select
	g.Phase = gesturefsm.GestPending
	g.EmptyDown = false
	g.DragNode = ""
	g.HandholdDown = false

	if h, ok := hitClassifiers[ev.Hit.Kind]; ok {
		h(md, g, ev)
	}
}

func (md *MoveDispatch) gestPointerMove(ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &md.UI.Gest
	if g.Phase == gesturefsm.GestIdle {
		return
	}
	dx := ev.X - g.DownX
	dy := ev.Y - g.DownY
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
	if g.Phase == gesturefsm.GestPending && dist > 0 && !g.Secondary {
		for _, edge := range commitEdges {
			if edge.guard(g) {
				edge.action(md, g, ev, tr)
				g.Phase = edge.to
				break
			}
		}
	}

	if apply, ok := applyAction[g.Phase]; ok {
		apply(md, g, ev, tr)
	}
}

func (md *MoveDispatch) gestPointerUp(ev inputcodec.RawInputMsg) {
	g := &md.UI.Gest
	switch {
	case g.Phase == gesturefsm.GestDragging:
		nodeGeoms, lq, ctx := md.MR.NodeGeoms(), &md.LQ, md.ctx
		applyNodeDragTarget(&md.UI, func(id string, target vec3) bool { return lq.RootMove(ctx, nodeGeoms, id, target) }, ev) // final target flush
	case g.Phase == gesturefsm.GestHandhold, g.Phase == gesturefsm.GestRotating:
		// Rotation completed (free or handhold-constrained): nothing to flush.
	case g.Phase == gesturefsm.GestPending:
		// Click → Go-owned selection. A node hit selects it; empty space clears the
		// selection. md.UI.Sel.Selected is the authoritative selection; Select() emits it so the
		// buffer snapshot marks the node's Selected column.
		md.applySelect(ev)
	}
	wasDragging := g.Phase == gesturefsm.GestDragging
	// Capture BEFORE reset() clears it (below) — movemsg.KindDragEnd must name the node
	// that was actually dragged, and reset() zeroes g.DragNode unconditionally.
	draggedNode := g.DragNode
	g.Reset(&md.UI.VP.Viewpoint)
	if wasDragging {
		// The drag just ended: g.DragNode is now "" (cleared by reset above), so the
		// Overlay block's DragNodeRow column must go back to -1 promptly rather than
		// waiting for the next unrelated view-frame emit. Mirrors commitDragStart's own
		// emitViewFrame call at drag START.
		md.UI.EmitViewFrame(nil)
		// "done dragging" (PLAN.md) — mirrors commitDragStart's own movemsg.KindDragStart
		// send. Sent on EVERY path a drag ends by (this is the FSM's one drag-end exit),
		// so a chain bead this node woke can never be left on machine time — see
		// movemsg.KindDragEnd's own doc comment.
		if draggedNode != "" {
			sendMove(&md.MR, md.ctx, draggedNode, movemsg.Msg{Kind: movemsg.KindDragEnd, NodeID: draggedNode})
		}
	}
}

// gestWheel mirrors interaction-handlers.ts handleWheelNative: ctrl+wheel = zoom-to-cursor
// dolly (expressed as a PAN in the polar model — a pivot translation, not a radius change),
// plain wheel = screen-space pan. Both first seed the viewpoint to region-focus, then pan.
func (md *MoveDispatch) gestWheel(ev inputcodec.RawInputMsg, tr *T.Trace) {
	vp := md.UI.VP.Viewpoint
	eye := geom.EyeOf(vp)
	pivot := geom.RegionFocus(vp, layoutquant.HeldCenters(md.MR.NodeGeoms(), md.MR.CenterOfNode))

	if ev.Ctrl {
		// Zoom-to-cursor: move the camera TOWARD the node under the cursor along the cursor→node
		// line, KEEPING the look direction — so that node stays fixed under the mouse. It does NOT
		// re-aim: re-aiming (snapping the camera to look straight at the node) is what recentered
		// the view and threw the cursor off. PanViewpoint translates the whole camera (pivot+eye
		// ride together); pos/up are unchanged, so the node keeps projecting to the same pixel.
		// The cursor→node pick is a screen-space selection at the input boundary (projectNDC).
		mouseNdcX, mouseNdcY := md.UI.Gest.PixelToNDC(ev.X, ev.Y)
		basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
		aspect := md.UI.Gest.Rect.Aspect()
		target := pivot
		best := math.Inf(1)
		for _, c := range layoutquant.HeldCenters(md.MR.NodeGeoms(), md.MR.CenterOfNode) {
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
		md.UI.VP.PanViewpoint(rayDir.Scale(step), tr)
		md.UI.EmitViewFrame(cameraViewEvent())
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
	worldPerPixel := (2 * vp.R * math.Tan(fovRad/2)) / md.UI.Gest.Rect.Height
	disp := geom.PanDisplacementPolar(vp.Pos, vp.Up, ev.DeltaX, ev.DeltaY, worldPerPixel)
	md.UI.VP.PanViewpoint(disp, tr)
	md.UI.EmitViewFrame(cameraViewEvent())
}
