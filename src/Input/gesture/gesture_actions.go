package gesture

import (
	"fmt"

	B "github.com/dtauraso/wirefold/src/Buffer"
	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Scene/rowtables"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
)

func updateHover(d Deps, ev inputcodec.RawInputMsg) {
	var node string
	switch ev.Hit.Kind {
	case "torus":

		if n, ok := d.RT.NodeFromHit(ev.Hit); ok {
			node = n
		}
	}
	mr, ctx := d.MR, d.Ctx
	sendMoveFn := func(id string, msg movemsg.Msg) { mr.SendMove(ctx, id, msg) }
	if setHover(d.UI, sendMoveFn, d.RT, node, "", false) {
		d.UI.EmitViewFrame(nil)
	}
}

func seedOrbitPivot(d Deps, pivot vec3) {
	vp := d.UI.VP.Viewpoint
	eye := Camera.EyeOf(vp)
	r := eye.Sub(pivot).Length()
	pos := Camera.WorldDirToAngles(eye.Sub(pivot))
	d.UI.VP.SetViewpoint(pivot, r, pos, vp.Up)
}

func applyOrbit(d Deps, ev inputcodec.RawInputMsg) {
	g := &d.UI.Gest
	vp := d.UI.VP.Viewpoint
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := Camera.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := Camera.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := Camera.ToWorldDir(basis, prev)
	currDir := Camera.ToWorldDir(basis, curr)
	d.UI.OrbitViewpoint(Camera.WorldDirToAngles(currDir), Camera.WorldDirToAngles(prevDir))
	after := d.UI.VP.Viewpoint
	d.UI.EmitBreadcrumb(B.RowEvent{
		Label: B.BreadcrumbOrbitStep, NodeRow: -1, PortRow: -1, TargetRow: -1,
		TargetPortRow: -1, EdgeRow: -1, Slot: -1, Value: 0,
		Text: fmt.Sprintf("prevxy=%.0f,%.0f xy=%.0f,%.0f cxy=%.0f,%.0f pxPerRad=%.2f posBefore=%.3f,%.3f posAfter=%.3f,%.3f r=%.1f",
			g.PrevX, g.PrevY, ev.X, ev.Y, g.RotCx, g.RotCy, g.RotPxPerRad,
			vp.Pos.Phi, vp.Pos.Theta, after.Pos.Phi, after.Pos.Theta, after.R),
	})
	d.UI.EmitViewFrame(nil)
}

func applyOrbitLocked(d Deps, ev inputcodec.RawInputMsg) {
	g := &d.UI.Gest
	vp := d.UI.VP.Viewpoint
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := Camera.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := Camera.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := Camera.ToWorldDir(basis, prev)
	currDir := Camera.ToWorldDir(basis, curr)
	d.UI.OrbitLockedViewpoint(Camera.WorldDirToAngles(currDir), Camera.WorldDirToAngles(prevDir))
	d.UI.EmitViewFrame(nil)
}

func applyNodeDragTarget(ui *viewstate.UIState, rootMove func(id string, target vec3) bool, ev inputcodec.RawInputMsg) bool {
	g := &ui.Gest
	hit, ok := ui.DragPlaneHit(ev)
	if !ok {
		return false
	}
	rootMove(g.DragNode, hit.Add(g.DragGrabOffset))
	return true
}

func setHover(ui *viewstate.UIState, sendMoveFn func(id string, msg movemsg.Msg), RT *rowtables.RowTables, node, port string, isInput bool) (changed bool) {
	if node == ui.Sel.HoverNode && port == ui.Sel.HoverPort && isInput == ui.Sel.HoverInput {
		return false
	}

	ui.SetHoverUI(sendMoveFn, node, port, isInput)
	return true
}

func applySelect(d Deps, ev inputcodec.RawInputMsg) {

	if ev.Hit.Kind == "empty" {
		setSelectionUI(d.UI, d.MR, d.Ctx, "", "")
		d.UI.EmitViewFrame(nil)
		return
	}
	if ev.Hit.Kind == "edge" {
		if label, ok := d.RT.EdgeFromHit(ev.Hit); ok {
			setSelectionUI(d.UI, d.MR, d.Ctx, "", label)

			d.UI.EmitViewFrame(nil)
			return
		}

	}

	var node string
	if ev.Hit.Kind == "node" {
		if n, ok := d.RT.NodeFromHit(ev.Hit); ok {
			node = n
		}
	}
	setSelectionUI(d.UI, d.MR, d.Ctx, node, "")
	d.UI.EmitViewFrame(nil)
}
