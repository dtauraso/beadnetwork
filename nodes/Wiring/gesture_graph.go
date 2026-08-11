package Wiring

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

// gesture_graph.go — an explicit adjacency-list TABLE + driver for the two gestPointerMove
// blocks (Block 1: pending→real-phase commit; Block 2: per-phase apply). This is a pure
// restructuring of the switches that used to live inline in gestPointerMove — same guards,
// same actions, same precedence order, same per-event sequencing (commit THEN apply, in the
// SAME gestPointerMove call). See gesture.go's gestPointerMove for the driver that walks
// this table.
//
// commitEdges is checked in order; the FIRST edge whose guard holds wins (mirrors the old
// switch{} arm order: wireNode, portMoveNode, dragNode, handholdDown, emptyDown). Its action
// runs, then the FSM phase is set to edge.to. applyAction is a per-phase table for Block 2
// (gestDragging/gestRotating/gestHandhold/gestPortMove) — a phase with no entry does nothing,
// same as the old switch's implicit default.

// gestureEdge is one entry of the pending→phase commit table.
type gestureEdge struct {
	guard  func(g *gestureState) bool
	action func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace)
	to     gesturePhase
}

// commitEdges is the SAME precedence order as the old commit switch in gestPointerMove:
// dragNode, handholdDown, emptyDown. wireNode/portMoveNode arms (gestWiring/gestPortMove)
// were removed with port geometry (docs/bead-model/channels-not-ports.md): a port is no longer drawn
// or hit-testable, so the "port" raycast-hit kind that fed both arms can never fire — they
// were dead code even before this deletion (wire-drop already created no edge; portMove's
// only effect was a ring-anchor snap that no longer exists).
var commitEdges = []gestureEdge{
	{
		guard: func(g *gestureState) bool { return g.DragNode != "" },
		action: func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace) {
			mr, ctx := &md.mr, md.ctx
			commitDragStart(&md.ui, func(id string, msg movemsg.Msg) { sendMove(mr, ctx, id, msg) }, g, ev, tr)
		},
		to: gestDragging,
	},
	{
		guard:  func(g *gestureState) bool { return g.HandholdDown },
		action: (*MoveDispatch).commitHandholdStart,
		to:     gestHandhold,
	},
	{
		guard:  func(g *gestureState) bool { return g.EmptyDown },
		action: (*MoveDispatch).commitRotateStart,
		to:     gestRotating,
	},
}

// commitDragStart is the drag-start commit action, copied verbatim from the old
// gestPointerMove commit switch's `case g.DragNode != ""` arm (minus the phase assignment,
// which the driver now performs from the edge's `to`). Side-effect order preserved:
// latch lastDraggedNode → sendMove(DragStart) (arms the dragged node's bead-actor wake,
// nodeMover.startBeadDrag). The local-polar drag-log reset
// (emitViewFrame(KindAbcDragReset) → resetAbcDrag()) that used to run here was deleted
// with the local-polar model itself (MODEL.md "the polar model") — there is no more
// per-drag recipient set to re-scope.
func commitDragStart(ui *uiState, sendMoveFn func(id string, msg movemsg.Msg), g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace) {
	// Capture the grab offset ONCE, at this exact slop-crossing commit event: the vector
	// from where the pointer's ray actually hits the drag plane to the node's center. Every
	// later move (applyNodeDragTarget) adds this same offset back so the grabbed point
	// stays under the cursor instead of the node's center jumping to it. Must happen here,
	// not inside the move path — see dragGrabOffset's doc comment in gesture.go. A parallel
	// ray (ok==false) leaves the offset at its zero value, degrading to centre-on-cursor
	// rather than breaking the drag.
	if hit, ok := ui.dragPlaneHit(ev); ok {
		g.DragGrabOffset = g.DragStartCenter.Sub(hit)
	}
	// Re-scope the in-editor drag-log to THIS drag. This is the ONE place a drag
	// begins (the slop-crossing pending→dragging transition), so it fires exactly
	// once per drag. It must NOT live in RootMove: that runs on every pointer-move
	// event of the drag, so resetting there interleaves with the neighborSetC fan's
	// AbcDrag marks (which land asynchronously on each recipient's own goroutine)
	// and drops recipients whose mark lands after the next move's reset.
	// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): this same goroutine also
	// writes its own VIEW frame directly, carrying this one-time drag-start event.
	// Latch this drag's node as the "last dragged" for the in-editor label, which must
	// persist past this drag's own pointerup (see uiState.lastDraggedNode's doc
	// comment). Set BEFORE the emitViewFrame call below, so that call's own
	// dragNodeRow derivation (which reads lastDraggedNode) already reflects this
	// drag, not the previous one.
	ui.lastDraggedNode = g.DragNode
	// Arm the dragged node's OWN bead-actor wake (movemsg.KindDragStart, see
	// nodeMover.startBeadDrag) at this same slop-crossing edge — the ONE place a drag
	// begins. Blocking send (sendMove, not lossy): this must not be dropped, same as
	// the drag/center kinds it rides alongside.
	sendMoveFn(g.DragNode, movemsg.Msg{Kind: movemsg.KindDragStart, NodeID: g.DragNode})
}

// commitHandholdStart is the handhold-constrained-orbit commit action, copied verbatim from
// the old commit switch's `case g.HandholdDown` arm (minus the phase assignment). Side-effect
// order preserved exactly: seed prevX/prevY and smoothX/smoothY from the grab point BEFORE
// seedOrbitPivot.
func (md *MoveDispatch) commitHandholdStart(g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace) {
	// Handhold-constrained orbit: seed prevX/prevY from the GRAB point (downX/downY),
	// not the slop-crossing point, so the first locked arc is grab→first-move (mirrors
	// interaction-handlers.ts). Seed the viewpoint about the frozen pivot, then lock.
	g.PrevX, g.PrevY = g.DownX, g.DownY
	g.SmoothX, g.SmoothY = g.DownX, g.DownY
	md.seedOrbitPivot(g.RotPivot)
}

// commitRotateStart is the empty-space-rotate commit action, copied verbatim from the old
// commit switch's `case g.EmptyDown` arm (minus the phase assignment). Side-effect order
// preserved exactly: seed prevX/prevY and smoothX/smoothY from the CURRENT point BEFORE
// seedOrbitPivot.
func (md *MoveDispatch) commitRotateStart(g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace) {
	g.PrevX, g.PrevY = ev.X, ev.Y
	g.SmoothX, g.SmoothY = ev.X, ev.Y
	// Seed the viewpoint so the orbit pivot is the frozen region-focus (mirrors the
	// TS sendViewpointSet at rotation start). pos/up/r recompute about the new pivot.
	md.seedOrbitPivot(g.RotPivot)
}

// applyAction is the Block 2 per-phase apply table, copied verbatim (per-case body) from the
// old `switch g.Phase` in gestPointerMove. A phase with no entry does nothing, matching the
// old switch's implicit default (gestIdle, gestPending fell through to nothing).
var applyAction = map[gesturePhase]func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace){
	gestDragging: func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace) {
		mr, lq, ctx := &md.mr, &md.lq, md.ctx
		if applyNodeDragTarget(&md.ui, func(id string, target vec3) bool { return lq.RootMove(ctx, mr, id, target) }, ev) {
			g.PrevX, g.PrevY = ev.X, ev.Y
		}
	},
	gestRotating: func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace) {
		g.SmoothX += geom.RotSmoothAlpha * (ev.X - g.SmoothX)
		g.SmoothY += geom.RotSmoothAlpha * (ev.Y - g.SmoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.SmoothX, g.SmoothY
		md.applyOrbit(smoothEv, tr)
		g.PrevX, g.PrevY = g.SmoothX, g.SmoothY
	},
	gestHandhold: func(md *MoveDispatch, g *gestureState, ev inputcodec.RawInputMsg, tr *T.Trace) {
		g.SmoothX += geom.RotSmoothAlpha * (ev.X - g.SmoothX)
		g.SmoothY += geom.RotSmoothAlpha * (ev.Y - g.SmoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.SmoothX, g.SmoothY
		md.applyOrbitLocked(smoothEv, tr)
		g.PrevX, g.PrevY = g.SmoothX, g.SmoothY
	},
}
