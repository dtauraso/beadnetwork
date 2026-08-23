package Gesture

import (
	"github.com/dtauraso/wirefold/Categories/Node"
	NodeGesture "github.com/dtauraso/wirefold/Categories/Node/Gesture"
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
	"github.com/dtauraso/wirefold/Categories/Scene/Drag"
)

func updateHover(d Deps, ev Drag.RawInputMsg) {
	var node string
	switch ev.Hit.Kind {
	case "torus":

		if n, ok := d.RT.NodeFromHit(ev.Hit); ok {
			node = n
		}
	}
	mr, ctx := d.MR, d.Ctx
	sendMoveFn := func(id string, msg Node.Msg) { mr.SendMove(ctx, id, msg) }
	if NodeGesture.SetHover(d.UI, sendMoveFn, node, "", false) {
		d.UI.EmitViewFrame(nil)
	}
}

func seedOrbitPivot(d Deps, pivot Vec3) {
	vp := d.UI.VP.Viewpoint
	eye := Camera.EyeOf(vp)
	r := eye.Sub(Camera.Vec3(pivot)).Length()
	pos := Camera.WorldDirToAngles(eye.Sub(Camera.Vec3(pivot)))
	d.UI.VP.SetViewpoint(Camera.Vec3(pivot), r, pos, vp.Up)
}

func applyOrbit(d Deps, ev Drag.RawInputMsg) {
	g := &d.UI.Gest
	vp := d.UI.VP.Viewpoint
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	prev := Camera.ScreenToPolar(g.PrevX-g.RotCx, g.PrevY-g.RotCy, g.RotPxPerRad)
	curr := Camera.ScreenToPolar(ev.X-g.RotCx, ev.Y-g.RotCy, g.RotPxPerRad)
	prevDir := Camera.ToWorldDir(basis, prev)
	currDir := Camera.ToWorldDir(basis, curr)
	d.UI.OrbitViewpoint(Camera.WorldDirToAngles(currDir), Camera.WorldDirToAngles(prevDir))
	d.UI.EmitViewFrame(nil)
}

func applyOrbitLocked(d Deps, ev Drag.RawInputMsg) {
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

func applySelect(d Deps, ev Drag.RawInputMsg) {

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
