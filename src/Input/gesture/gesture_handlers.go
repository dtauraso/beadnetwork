package gesture

import (
	"github.com/dtauraso/wirefold/src/Chrome/Pills/FitButton"
	"math"

	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Input/gesturefsm"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/movemsg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodemove"
)

func gestHome(d Deps, ev inputcodec.RawInputMsg) {
	centers := nodemove.HeldCenters(d.MR.NodeGeoms(), d.MR.CenterOfNode)
	radius := make(map[string]float64, len(centers))
	for id := range centers {
		radius[id] = d.MR.NodeBodyRadius(id)
	}
	pivot, r, pos, up, ok := FitButton.HomeFitPose(centers, radius, d.UI.FovDeg(), d.UI.Gest.Rect.Aspect())
	if !ok {
		return
	}
	d.UI.VP.SetViewpoint(pivot, r, pos, up)
	d.UI.VP.EmitViewpoint()
	d.UI.EmitViewFrame(nil)
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

func gestPointerMove(d Deps, ev inputcodec.RawInputMsg) {
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
		apply(d, g, ev)
	}
}

func gestPointerUp(d Deps, ev inputcodec.RawInputMsg) {
	g := &d.UI.Gest
	switch {
	case g.Phase == gesturefsm.GestDragging:
		nodeGeoms, mv, ctx := d.MR.NodeGeoms(), d.Mover, d.Ctx
		applyNodeDragTarget(d.UI, func(id string, target vec3) bool { return mv.RootMove(ctx, nodeGeoms, id, target) }, ev)
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

func gestWheel(d Deps, ev inputcodec.RawInputMsg) {
	vp := d.UI.VP.Viewpoint
	eye := Camera.EyeOf(vp)
	pivot := Camera.RegionFocus(vp, nodemove.HeldCenters(d.MR.NodeGeoms(), d.MR.CenterOfNode))

	if ev.Ctrl {

		mouseNdcX, mouseNdcY := d.UI.Gest.PixelToNDC(ev.X, ev.Y)
		basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
		aspect := d.UI.Gest.Rect.Aspect()
		target := pivot
		best := math.Inf(1)
		for _, c := range nodemove.HeldCenters(d.MR.NodeGeoms(), d.MR.CenterOfNode) {
			nx, ny, inFront := Camera.ProjectNDC(c, eye, basis, d.UI.FovDeg(), aspect)
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
		rayDir := Camera.AnglesToWorldOffset(1, vp.Pos.Phi, vp.Pos.Theta).Scale(-1)
		if distP > 1e-9 {
			rayDir = toTarget.Scale(1 / distP)
		}

		amt := 1 - math.Pow(Camera.GestureZoomBase, ev.DeltaY)
		step := distP * amt
		if minStep := vp.R * (Camera.GestureZoomBase - 1); math.Abs(step) < minStep {
			step = math.Copysign(minStep, amt)
		}
		d.UI.VP.PanViewpoint(rayDir.Scale(step))
		d.UI.EmitViewFrame(nil)
		return
	}

	fovRad := d.UI.FovDeg() * math.Pi / 180
	worldPerPixel := (2 * vp.R * math.Tan(fovRad/2)) / d.UI.Gest.Rect.Height
	disp := Camera.PanDisplacementPolar(vp.Pos, vp.Up, ev.DeltaX, ev.DeltaY, worldPerPixel)
	d.UI.VP.PanViewpoint(disp)
	d.UI.EmitViewFrame(nil)
}
