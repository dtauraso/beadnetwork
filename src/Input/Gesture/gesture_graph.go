package Gesture

import (
	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Input/Fsm"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
	"github.com/dtauraso/wirefold/src/spatial"
)

type gestureEdge struct {
	guard  func(g *Fsm.GestureState) bool
	action func(d Deps, g *Fsm.GestureState, ev Codec.RawInputMsg)
	to     Fsm.GesturePhase
}

var commitEdges = []gestureEdge{
	{
		guard: func(g *Fsm.GestureState) bool { return g.DragNode != "" },
		action: func(d Deps, g *Fsm.GestureState, ev Codec.RawInputMsg) {
			mr, ctx := d.MR, d.Ctx
			commitDragStart(d.UI, func(id string, msg movemsg.Msg) { mr.SendMove(ctx, id, msg) }, g, ev)
		},
		to: Fsm.GestDragging,
	},
	{
		guard:  func(g *Fsm.GestureState) bool { return g.HandholdDown },
		action: commitHandholdStart,
		to:     Fsm.GestHandhold,
	},
	{
		guard:  func(g *Fsm.GestureState) bool { return g.EmptyDown },
		action: commitRotateStart,
		to:     Fsm.GestRotating,
	},
}

func commitDragStart(ui *viewstate.UIState, sendMoveFn func(id string, msg movemsg.Msg), g *Fsm.GestureState, ev Codec.RawInputMsg) {

	if hit, ok := ui.DragPlaneHit(ev); ok {
		g.DragGrabOffset = g.DragStartCenter.Sub(hit)
	}

	ui.LastDraggedNode = g.DragNode

	sendMoveFn(g.DragNode, movemsg.Msg{Kind: movemsg.KindDragStart, NodeID: g.DragNode})
}

func commitHandholdStart(d Deps, g *Fsm.GestureState, ev Codec.RawInputMsg) {

	g.PrevX, g.PrevY = g.DownX, g.DownY
	g.SmoothX, g.SmoothY = g.DownX, g.DownY
	seedOrbitPivot(d, g.RotPivot)
}

func commitRotateStart(d Deps, g *Fsm.GestureState, ev Codec.RawInputMsg) {
	g.PrevX, g.PrevY = ev.X, ev.Y
	g.SmoothX, g.SmoothY = ev.X, ev.Y

	seedOrbitPivot(d, g.RotPivot)
}

var applyAction = map[Fsm.GesturePhase]func(d Deps, g *Fsm.GestureState, ev Codec.RawInputMsg){
	Fsm.GestDragging: func(d Deps, g *Fsm.GestureState, ev Codec.RawInputMsg) {
		nodeGeoms, mv, ctx := d.MR.NodeGeoms(), d.Mover, d.Ctx
		if applyNodeDragTarget(d.UI, func(id string, target spatial.Vec3) bool { return mv.RootMove(ctx, nodeGeoms, id, target) }, ev) {
			g.PrevX, g.PrevY = ev.X, ev.Y
		}
	},
	Fsm.GestRotating: func(d Deps, g *Fsm.GestureState, ev Codec.RawInputMsg) {
		g.SmoothX += Camera.RotSmoothAlpha * (ev.X - g.SmoothX)
		g.SmoothY += Camera.RotSmoothAlpha * (ev.Y - g.SmoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.SmoothX, g.SmoothY
		applyOrbit(d, smoothEv)
		g.PrevX, g.PrevY = g.SmoothX, g.SmoothY
	},
	Fsm.GestHandhold: func(d Deps, g *Fsm.GestureState, ev Codec.RawInputMsg) {
		g.SmoothX += Camera.RotSmoothAlpha * (ev.X - g.SmoothX)
		g.SmoothY += Camera.RotSmoothAlpha * (ev.Y - g.SmoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.SmoothX, g.SmoothY
		applyOrbitLocked(d, smoothEv)
		g.PrevX, g.PrevY = g.SmoothX, g.SmoothY
	},
}
