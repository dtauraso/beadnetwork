package Gesture

import (
	"fmt"
	"math"

	T "github.com/dtauraso/wirefold/src/Trace"
	"github.com/dtauraso/wirefold/src/spatial"

	"github.com/dtauraso/wirefold/src/Chrome/Pills/FitButton"

	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Input/Fsm"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/nodemove"
)

func gestHome(d Deps, ev Codec.RawInputMsg) {
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

func gestPointerDown(d Deps, ev Codec.RawInputMsg) {
	g := &d.UI.Gest
	g.DownX, g.DownY = ev.X, ev.Y
	g.PrevX, g.PrevY = ev.X, ev.Y
	g.Button = ev.Button
	g.Secondary = ev.Button == 2
	g.Phase = Fsm.GestPending
	g.EmptyDown = false
	g.DragNode = ""
	g.HandholdDown = false

	if h, ok := hitClassifiers[ev.Hit.Kind]; ok {
		h(d, g, ev)
	}

	d.UI.EmitBreadcrumb(T.RowEvent{
		Label: T.BreadcrumbPointerDown, NodeRow: -1, PortRow: -1, TargetRow: -1,
		TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(ev.Button),
		Text: fmt.Sprintf("hit=%q empty=%t handhold=%t node=%q xy=%.0f,%.0f rect=%.0f,%.0f,%.0fx%.0f pxPerRad=%.2f",
			ev.Hit.Kind, g.EmptyDown, g.HandholdDown, g.DragNode,
			ev.X, ev.Y, g.Rect.Left, g.Rect.Top, g.Rect.Width, g.Rect.Height, g.RotPxPerRad),
	})
}

func gestPointerMove(d Deps, ev Codec.RawInputMsg) {
	g := &d.UI.Gest
	if g.Phase == Fsm.GestIdle {
		return
	}
	dx := ev.X - g.DownX
	dy := ev.Y - g.DownY
	dist := math.Hypot(dx, dy)

	if g.Phase == Fsm.GestPending && dist > 0 && !g.Secondary {
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

func gestPointerUp(d Deps, ev Codec.RawInputMsg) {
	g := &d.UI.Gest
	switch {
	case g.Phase == Fsm.GestDragging:
		nodeGeoms, mv, ctx := d.MR.NodeGeoms(), d.Mover, d.Ctx
		applyNodeDragTarget(d.UI, func(id string, target spatial.Vec3) bool { return mv.RootMove(ctx, nodeGeoms, id, target) }, ev)
	case g.Phase == Fsm.GestHandhold, g.Phase == Fsm.GestRotating:

	case g.Phase == Fsm.GestPending:

		applySelect(d, ev)
	}
	wasDragging := g.Phase == Fsm.GestDragging

	draggedNode := g.DragNode
	g.Reset(&d.UI.VP.Viewpoint)
	if wasDragging {

		d.UI.EmitViewFrame(nil)

		if draggedNode != "" {
			d.MR.SendMove(d.Ctx, draggedNode, movemsg.Msg{Kind: movemsg.KindDragEnd, NodeID: draggedNode})
		}
	}
}

func gestWheel(d Deps, ev Codec.RawInputMsg) {
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
