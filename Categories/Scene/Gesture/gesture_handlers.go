package Gesture

import (
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Node"
	NodeGesture "github.com/dtauraso/beadnetwork/Categories/Node/Gesture"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/FitButton"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
)

func gestHome(d Deps, ev Drag.RawInputMsg) {
	centers := heldCenters(d)
	radius := make(map[string]float64, len(centers))
	for id := range centers {
		radius[id] = d.MR.BodyRadius(id)
	}
	pivot, r, pos, up, ok := FitButton.HomeFitPose(centersForFit(centers), radius, d.UI.FovDeg(), d.UI.Gest.Rect.Aspect())
	if !ok {
		return
	}
	d.UI.VP.SetViewpoint(Camera.Vec3(pivot), r, pos, up)
	d.UI.VP.EmitViewpoint()
	d.UI.EmitViewFrame(nil)
}

func gestPointerDown(d Deps, ev Drag.RawInputMsg) {
	g := &d.UI.Gest
	g.DownX, g.DownY = ev.X, ev.Y
	g.PrevX, g.PrevY = ev.X, ev.Y
	g.Button = ev.Button
	g.Secondary = ev.Button == 2
	g.Phase = Drag.GestPending
	g.EmptyDown = false
	g.NodeDrag.Clear()
	g.HandholdDown = false

	if h, ok := beginByHitKind[ev.Hit.Kind]; ok {
		h(d, g, ev)
	}

	d.UI.EmitBreadcrumb(View.RowEvent{
		Label: BreadcrumbPointerDown, NodeRow: -1, PortRow: -1, TargetRow: -1,
		TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(ev.Button),
		Text: fmt.Sprintf("hit=%q empty=%t handhold=%t node=%q xy=%.0f,%.0f rect=%.0f,%.0f,%.0fx%.0f ball=%.1f,%.1f,%.1f",
			ev.Hit.Kind, g.EmptyDown, g.HandholdDown, g.NodeDrag.Node,
			ev.X, ev.Y, g.Rect.Left, g.Rect.Top, g.Rect.Width, g.Rect.Height, ev.Ball.X, ev.Ball.Y, ev.Ball.Z),
	})
}

func gestPointerMove(d Deps, ev Drag.RawInputMsg) {
	g := &d.UI.Gest
	if g.Phase == Drag.GestIdle {
		return
	}
	dx := ev.X - g.DownX
	dy := ev.Y - g.DownY
	dist := math.Hypot(dx, dy)

	if g.Phase == Drag.GestPending && dist > 0 && !g.Secondary {
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

func gestPointerUp(d Deps, ev Drag.RawInputMsg) {
	g := &d.UI.Gest
	switch {
	case g.Phase == Drag.GestDragging:
		nodeGeoms, mv, ctx := d.MR.NodeGeoms(), d.Mover, d.Ctx
		NodeGesture.ApplyDragTarget(d.UI, func(id string, target Node.Vec3) bool { return mv.RootMove(ctx, nodeGeoms, id, target) }, ev)
	case g.Phase == Drag.GestHandhold, g.Phase == Drag.GestRotating:

	case g.Phase == Drag.GestPending:

		applySelect(d, ev)
	}
	wasDragging := g.Phase == Drag.GestDragging

	draggedNode := g.NodeDrag.Node
	g.Reset(&d.UI.VP.Viewpoint)
	if wasDragging {

		d.UI.EmitViewFrame(nil)

		if draggedNode != "" {
			d.MR.SendMove(d.Ctx, draggedNode, Node.Msg{NodeID: draggedNode, Body: Node.DragEnd{}})
		}
	}
}

func gestWheel(d Deps, ev Drag.RawInputMsg) {
	vp := d.UI.VP.Viewpoint
	eye := Camera.EyeOf(vp)
	pivot := Camera.RegionFocus(vp, centersForCamera(heldCenters(d)))

	if ev.Ctrl {
		target := pivot
		if node, ok := d.RT.NodeFromHit(ev.Hit); ok {
			if c, ok := d.MR.CenterOf(node); ok {
				target = Camera.Vec3(c)
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
