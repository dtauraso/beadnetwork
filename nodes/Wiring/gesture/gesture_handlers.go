package gesture

import (
	"math"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

func gestHome(d Deps, ev inputcodec.RawInputMsg, tr *T.Trace) {
	centers := layoutquant.HeldCenters(d.MR.NodeGeoms(), d.MR.CenterOfNode)
	radius := make(map[string]float64, len(centers))
	for id := range centers {
		radius[id] = d.MR.NodeBodyRadius(id)
	}
	pivot, r, pos, up, ok := camera.HomeFitPose(centers, radius, ev.Fov, d.UI.Gest.Rect.Aspect())
	if !ok {
		return
	}
	d.UI.VP.SetViewpoint(pivot, r, pos, up)
	d.UI.VP.EmitViewpoint(tr)
	d.UI.EmitViewFrame(CameraViewEvent())
}

func gestPointerDown(d Deps, ev inputcodec.RawInputMsg) {
	g := &d.UI.Gest
	g.DownX, g.DownY = ev.X, ev.Y
	g.PrevX, g.PrevY = ev.X, ev.Y
	g.Button = ev.Button
	g.Secondary = ev.Button == 2
	g.Phase = gesturefsm.GestPending
	g.EmptyDown = false
	g.DragNode = ""
	g.HandholdDown = false

	if h, ok := hitClassifiers[ev.Hit.Kind]; ok {
		h(d, g, ev)
	}
}

func gestPointerMove(d Deps, ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &d.UI.Gest
	if g.Phase == gesturefsm.GestIdle {
		return
	}
	dx := ev.X - g.DownX
	dy := ev.Y - g.DownY
	dist := math.Hypot(dx, dy)

	if g.Phase == gesturefsm.GestPending && dist > 0 && !g.Secondary {
		for _, edge := range commitEdges {
			if edge.guard(g) {
				edge.action(d, g, ev)
				g.Phase = edge.to
				break
			}
		}
	}

	if apply, ok := applyAction[g.Phase]; ok {
		apply(d, g, ev, tr)
	}
}

func gestPointerUp(d Deps, ev inputcodec.RawInputMsg) {
	g := &d.UI.Gest
	switch {
	case g.Phase == gesturefsm.GestDragging:
		nodeGeoms, lq, ctx := d.MR.NodeGeoms(), d.LQ, d.Ctx
		applyNodeDragTarget(d.UI, func(id string, target vec3) bool { return lq.RootMove(ctx, nodeGeoms, d.MR.CenterOfNode, id, target) }, ev)
	case g.Phase == gesturefsm.GestHandhold, g.Phase == gesturefsm.GestRotating:

	case g.Phase == gesturefsm.GestPending:

		applySelect(d, ev)
	}
	wasDragging := g.Phase == gesturefsm.GestDragging

	draggedNode := g.DragNode
	g.Reset(&d.UI.VP.Viewpoint)
	if wasDragging {

		d.UI.EmitViewFrame(nil)

		if draggedNode != "" {
			d.MR.SendMove(d.Ctx, draggedNode, movemsg.Msg{Kind: movemsg.KindDragEnd, NodeID: draggedNode})
		}
	}
}

func gestWheel(d Deps, ev inputcodec.RawInputMsg, tr *T.Trace) {
	vp := d.UI.VP.Viewpoint
	eye := camera.EyeOf(vp)
	pivot := camera.RegionFocus(vp, layoutquant.HeldCenters(d.MR.NodeGeoms(), d.MR.CenterOfNode))

	if ev.Ctrl {

		mouseNdcX, mouseNdcY := d.UI.Gest.PixelToNDC(ev.X, ev.Y)
		basis := camera.BasisFromViewpoint(vp.Pos, vp.Up)
		aspect := d.UI.Gest.Rect.Aspect()
		target := pivot
		best := math.Inf(1)
		for _, c := range layoutquant.HeldCenters(d.MR.NodeGeoms(), d.MR.CenterOfNode) {
			nx, ny, inFront := camera.ProjectNDC(c, eye, basis, ev.Fov, aspect)
			if !inFront {
				continue
			}
			if dd := math.Hypot(nx-mouseNdcX, ny-mouseNdcY); dd < best {
				best = dd
				target = c
			}
		}
		toTarget := target.Sub(eye)
		distP := toTarget.Length()
		rayDir := camera.AnglesToWorldOffset(1, vp.Pos.Phi, vp.Pos.Theta).Scale(-1)
		if distP > 1e-9 {
			rayDir = toTarget.Scale(1 / distP)
		}

		amt := 1 - math.Pow(camera.GestureZoomBase, ev.DeltaY)
		step := distP * amt
		if minStep := vp.R * (camera.GestureZoomBase - 1); math.Abs(step) < minStep {
			step = math.Copysign(minStep, amt)
		}
		d.UI.VP.PanViewpoint(rayDir.Scale(step), tr)
		d.UI.EmitViewFrame(CameraViewEvent())
		return
	}

	fovRad := ev.Fov * math.Pi / 180
	worldPerPixel := (2 * vp.R * math.Tan(fovRad/2)) / d.UI.Gest.Rect.Height
	disp := camera.PanDisplacementPolar(vp.Pos, vp.Up, ev.DeltaX, ev.DeltaY, worldPerPixel)
	d.UI.VP.PanViewpoint(disp, tr)
	d.UI.EmitViewFrame(CameraViewEvent())
}
