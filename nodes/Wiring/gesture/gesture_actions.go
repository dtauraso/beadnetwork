package gesture

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"

	T "github.com/dtauraso/wirefold/tools/topology-vscode/src/Trace"
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
	eye := camera.EyeOf(vp)
	r := eye.Sub(pivot).Length()
	pos := camera.WorldDirToAngles(eye.Sub(pivot))
	d.UI.VP.SetViewpoint(pivot, r, pos, vp.Up)
}

func applyOrbit(d Deps, ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &d.UI.Gest
	vp := d.UI.VP.Viewpoint
	basis := camera.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := camera.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := camera.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := camera.ToWorldDir(basis, prev)
	currDir := camera.ToWorldDir(basis, curr)
	d.UI.OrbitViewpoint(camera.WorldDirToAngles(currDir), camera.WorldDirToAngles(prevDir), tr)
	d.UI.EmitViewFrame(nil)
}

func applyOrbitLocked(d Deps, ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &d.UI.Gest
	vp := d.UI.VP.Viewpoint
	basis := camera.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := camera.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := camera.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := camera.ToWorldDir(basis, prev)
	currDir := camera.ToWorldDir(basis, curr)
	d.UI.OrbitLockedViewpoint(camera.WorldDirToAngles(currDir), camera.WorldDirToAngles(prevDir), tr)
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
