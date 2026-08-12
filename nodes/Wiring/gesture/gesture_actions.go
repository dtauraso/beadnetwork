package gesture

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
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
	if events, changed := setHover(d.UI, sendMoveFn, d.RT, node, "", false); changed {
		d.UI.EmitViewFrame(events)
	}
}

func seedOrbitPivot(d Deps, pivot vec3) {
	vp := d.UI.VP.Viewpoint
	eye := geom.EyeOf(vp)
	r := eye.Sub(pivot).Length()
	pos := geom.WorldDirToAngles(eye.Sub(pivot))
	d.UI.VP.SetViewpoint(pivot, r, pos, vp.Up)
}

func applyOrbit(d Deps, ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &d.UI.Gest
	vp := d.UI.VP.Viewpoint
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := geom.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := geom.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := geom.ToWorldDir(basis, prev)
	currDir := geom.ToWorldDir(basis, curr)
	d.UI.OrbitViewpoint(geom.WorldDirToAngles(currDir), geom.WorldDirToAngles(prevDir), tr)
	d.UI.EmitViewFrame(CameraViewEvent())
}

func applyOrbitLocked(d Deps, ev inputcodec.RawInputMsg, tr *T.Trace) {
	g := &d.UI.Gest
	vp := d.UI.VP.Viewpoint
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := geom.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := geom.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := geom.ToWorldDir(basis, prev)
	currDir := geom.ToWorldDir(basis, curr)
	d.UI.OrbitLockedViewpoint(geom.WorldDirToAngles(currDir), geom.WorldDirToAngles(prevDir), tr)
	d.UI.EmitViewFrame(CameraViewEvent())
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

func setHover(ui *viewstate.UIState, sendMoveFn func(id string, msg movemsg.Msg), RT *rowtables.RowTables, node, port string, isInput bool) (events []rowevent.RowEvent, changed bool) {
	if node == ui.Sel.HoverNode && port == ui.Sel.HoverPort && isInput == ui.Sel.HoverInput {
		return nil, false
	}

	ui.SetHoverUI(sendMoveFn, node, port, isInput)
	nodeRow := int32(-1)
	if r, ok := RT.NodeRowFor(node); ok {
		nodeRow = r
	}

	portRow := int32(-1)
	value := int32(0)
	if isInput {
		value = 1
	}
	return []rowevent.RowEvent{{Kind: T.KindHover, NodeRow: nodeRow, PortRow: portRow, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: value}}, true
}

func applySelect(d Deps, ev inputcodec.RawInputMsg) {

	if ev.Hit.Kind == "empty" {
		setSelectionUI(d.UI, d.MR, d.Ctx, "", "")
		d.UI.EmitViewFrame(d.RT.SelectViewEvent(""))
		return
	}
	if ev.Hit.Kind == "edge" {
		if label, ok := d.RT.EdgeFromHit(ev.Hit); ok {
			setSelectionUI(d.UI, d.MR, d.Ctx, "", label)

			d.UI.EmitViewFrame(d.RT.SelectViewEvent(""))
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
	d.UI.EmitViewFrame(d.RT.SelectViewEvent(node))
}
