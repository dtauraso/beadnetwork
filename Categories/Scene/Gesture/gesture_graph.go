package Gesture

import (
	"github.com/dtauraso/wirefold/Categories/Node"
	NodeGesture "github.com/dtauraso/wirefold/Categories/Node/Gesture"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
	"github.com/dtauraso/wirefold/Categories/Scene/Drag"
)

type gestureEdge struct {
	guard  func(g *Drag.GestureState) bool
	action func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg)
	to     Drag.GesturePhase
}

var commitEdges = []gestureEdge{
	{
		guard: func(g *Drag.GestureState) bool { return g.NodeDrag.Holding() },
		action: func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
			mr, ctx := d.MR, d.Ctx
			NodeGesture.CommitDragStart(d.UI, func(id string, msg Node.Msg) { mr.SendMove(ctx, id, msg) }, g, ev)
		},
		to: Drag.GestDragging,
	},
	{
		guard:  func(g *Drag.GestureState) bool { return g.HandholdDown },
		action: commitHandholdStart,
		to:     Drag.GestHandhold,
	},
	{
		guard:  func(g *Drag.GestureState) bool { return g.EmptyDown },
		action: commitRotateStart,
		to:     Drag.GestRotating,
	},
}

func commitHandholdStart(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {

	g.PrevX, g.PrevY = g.DownX, g.DownY
	g.SmoothX, g.SmoothY = g.DownX, g.DownY
	seedOrbitPivot(d, Vec3(g.RotPivot))
}

func commitRotateStart(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
	g.PrevX, g.PrevY = ev.X, ev.Y
	g.SmoothX, g.SmoothY = ev.X, ev.Y

	seedOrbitPivot(d, Vec3(g.RotPivot))
}

var applyAction = map[Drag.GesturePhase]func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg){
	Drag.GestDragging: func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		nodeGeoms, mv, ctx := d.MR.NodeGeoms(), d.Mover, d.Ctx
		if NodeGesture.ApplyDragTarget(d.UI, func(id string, target nodegeom.Vec3) bool { return mv.RootMove(ctx, nodeGeoms, id, target) }, ev) {
			g.PrevX, g.PrevY = ev.X, ev.Y
		}
	},
	Drag.GestRotating: func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		g.SmoothX += Camera.RotSmoothAlpha * (ev.X - g.SmoothX)
		g.SmoothY += Camera.RotSmoothAlpha * (ev.Y - g.SmoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.SmoothX, g.SmoothY
		applyOrbit(d, smoothEv)
		g.PrevX, g.PrevY = g.SmoothX, g.SmoothY
	},
	Drag.GestHandhold: func(d Deps, g *Drag.GestureState, ev Drag.RawInputMsg) {
		g.SmoothX += Camera.RotSmoothAlpha * (ev.X - g.SmoothX)
		g.SmoothY += Camera.RotSmoothAlpha * (ev.Y - g.SmoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.SmoothX, g.SmoothY
		applyOrbitLocked(d, smoothEv)
		g.PrevX, g.PrevY = g.SmoothX, g.SmoothY
	},
}
