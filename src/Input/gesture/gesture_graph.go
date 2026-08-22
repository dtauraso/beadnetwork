package gesture

import (
	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Input/gesturefsm"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
	"github.com/dtauraso/wirefold/src/spatial"
)

type gestureEdge struct {
	guard  func(g *gesturefsm.GestureState) bool
	action func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg)
	to     gesturefsm.GesturePhase
}

var commitEdges = []gestureEdge{
	{
		guard: func(g *gesturefsm.GestureState) bool { return g.DragNode != "" },
		action: func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
			mr, ctx := d.MR, d.Ctx
			commitDragStart(d.UI, func(id string, msg movemsg.Msg) { mr.SendMove(ctx, id, msg) }, g, ev)
		},
		to: gesturefsm.GestDragging,
	},
	{
		guard:  func(g *gesturefsm.GestureState) bool { return g.HandholdDown },
		action: commitHandholdStart,
		to:     gesturefsm.GestHandhold,
	},
	{
		guard:  func(g *gesturefsm.GestureState) bool { return g.EmptyDown },
		action: commitRotateStart,
		to:     gesturefsm.GestRotating,
	},
}

func commitDragStart(ui *viewstate.UIState, sendMoveFn func(id string, msg movemsg.Msg), g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {

	if hit, ok := ui.DragPlaneHit(ev); ok {
		g.DragGrabOffset = g.DragStartCenter.Sub(hit)
	}

	ui.LastDraggedNode = g.DragNode

	sendMoveFn(g.DragNode, movemsg.Msg{Kind: movemsg.KindDragStart, NodeID: g.DragNode})
}

func commitHandholdStart(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {

	g.PrevX, g.PrevY = g.DownX, g.DownY
	g.SmoothX, g.SmoothY = g.DownX, g.DownY
	seedOrbitPivot(d, g.RotPivot)
}

func commitRotateStart(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
	g.PrevX, g.PrevY = ev.X, ev.Y
	g.SmoothX, g.SmoothY = ev.X, ev.Y

	seedOrbitPivot(d, g.RotPivot)
}

var applyAction = map[gesturefsm.GesturePhase]func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg){
	gesturefsm.GestDragging: func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
		nodeGeoms, mv, ctx := d.MR.NodeGeoms(), d.Mover, d.Ctx
		if applyNodeDragTarget(d.UI, func(id string, target spatial.Vec3) bool { return mv.RootMove(ctx, nodeGeoms, id, target) }, ev) {
			g.PrevX, g.PrevY = ev.X, ev.Y
		}
	},
	gesturefsm.GestRotating: func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
		g.SmoothX += Camera.RotSmoothAlpha * (ev.X - g.SmoothX)
		g.SmoothY += Camera.RotSmoothAlpha * (ev.Y - g.SmoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.SmoothX, g.SmoothY
		applyOrbit(d, smoothEv)
		g.PrevX, g.PrevY = g.SmoothX, g.SmoothY
	},
	gesturefsm.GestHandhold: func(d Deps, g *gesturefsm.GestureState, ev inputcodec.RawInputMsg) {
		g.SmoothX += Camera.RotSmoothAlpha * (ev.X - g.SmoothX)
		g.SmoothY += Camera.RotSmoothAlpha * (ev.Y - g.SmoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.SmoothX, g.SmoothY
		applyOrbitLocked(d, smoothEv)
		g.PrevX, g.PrevY = g.SmoothX, g.SmoothY
	},
}
