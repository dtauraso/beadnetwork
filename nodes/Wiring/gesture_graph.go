package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
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
	action func(md *MoveDispatch, g *gestureState, ev rawInputMsg, tr *T.Trace)
	to     gesturePhase
}

// commitEdges is the SAME precedence order as the old commit switch in gestPointerMove:
// wireNode, portMoveNode, dragNode, handholdDown, emptyDown.
var commitEdges = []gestureEdge{
	{
		guard:  func(g *gestureState) bool { return g.wireNode != "" },
		action: func(md *MoveDispatch, g *gestureState, ev rawInputMsg, tr *T.Trace) {},
		to:     gestWiring,
	},
	{
		guard:  func(g *gestureState) bool { return g.portMoveNode != "" },
		action: func(md *MoveDispatch, g *gestureState, ev rawInputMsg, tr *T.Trace) {},
		to:     gestPortMove,
	},
	{
		guard:  func(g *gestureState) bool { return g.dragNode != "" },
		action: (*MoveDispatch).commitDragStart,
		to:     gestDragging,
	},
	{
		guard:  func(g *gestureState) bool { return g.handholdDown },
		action: (*MoveDispatch).commitHandholdStart,
		to:     gestHandhold,
	},
	{
		guard:  func(g *gestureState) bool { return g.emptyDown },
		action: (*MoveDispatch).commitRotateStart,
		to:     gestRotating,
	},
}

// commitDragStart is the drag-start commit action, copied verbatim from the old
// gestPointerMove commit switch's `case g.dragNode != ""` arm (minus the phase assignment,
// which the driver now performs from the edge's `to`). Side-effect order preserved exactly:
// emitViewFrame(KindAbcDragReset) → resetAbcDrag() → sendMove(DragStart).
func (md *MoveDispatch) commitDragStart(g *gestureState, ev rawInputMsg, tr *T.Trace) {
	// Re-scope the in-editor drag-log to THIS drag. This is the ONE place a drag
	// begins (the slop-crossing pending→dragging transition), so it fires exactly
	// once per drag. It must NOT live in RootMove: that runs on every pointer-move
	// event of the drag, so resetting there interleaves with the neighborSetC fan's
	// AbcDrag marks (which land asynchronously on each recipient's own goroutine)
	// and drops recipients whose mark lands after the next move's reset.
	// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): this same goroutine also
	// writes its own VIEW frame directly, carrying this one-time drag-start event.
	md.emitViewFrame([]wire.RowEvent{{Kind: T.KindAbcDragReset, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}})
	// Re-scope MoveDispatch's OWN published recipient set the same way (count is
	// a cumulative total-events affirmation and is intentionally left alone — only
	// the NAME SET is drag-scoped, mirroring the old central accumulator's
	// KindAbcDragReset handling of its abcDragged/gotDragMsg).
	md.resetAbcDrag()
	// Arm the dragged node's OWN drag-anchor snapshot (moveMsgKindDragStart, see
	// its doc comment in node_move.go) at this same slop-crossing edge — the ONE
	// place a drag begins — so the in-editor delta log reads the drag's running
	// total from this exact start point instead of a per-move-event (0,0,0).
	// Blocking send (md.sendMove, not lossy): this must not be dropped, same as
	// the drag/center kinds it rides alongside.
	md.sendMove(g.dragNode, moveMsg{Kind: moveMsgKindDragStart, NodeID: g.dragNode})
}

// commitHandholdStart is the handhold-constrained-orbit commit action, copied verbatim from
// the old commit switch's `case g.handholdDown` arm (minus the phase assignment). Side-effect
// order preserved exactly: seed prevX/prevY and smoothX/smoothY from the grab point BEFORE
// seedOrbitPivot.
func (md *MoveDispatch) commitHandholdStart(g *gestureState, ev rawInputMsg, tr *T.Trace) {
	// Handhold-constrained orbit: seed prevX/prevY from the GRAB point (downX/downY),
	// not the slop-crossing point, so the first locked arc is grab→first-move (mirrors
	// interaction-handlers.ts). Seed the viewpoint about the frozen pivot, then lock.
	g.prevX, g.prevY = g.downX, g.downY
	g.smoothX, g.smoothY = g.downX, g.downY
	md.seedOrbitPivot(g.rotPivot)
}

// commitRotateStart is the empty-space-rotate commit action, copied verbatim from the old
// commit switch's `case g.emptyDown` arm (minus the phase assignment). Side-effect order
// preserved exactly: seed prevX/prevY and smoothX/smoothY from the CURRENT point BEFORE
// seedOrbitPivot.
func (md *MoveDispatch) commitRotateStart(g *gestureState, ev rawInputMsg, tr *T.Trace) {
	g.prevX, g.prevY = ev.X, ev.Y
	g.smoothX, g.smoothY = ev.X, ev.Y
	// Seed the viewpoint so the orbit pivot is the frozen region-focus (mirrors the
	// TS sendViewpointSet at rotation start). pos/up/r recompute about the new pivot.
	md.seedOrbitPivot(g.rotPivot)
}

// applyAction is the Block 2 per-phase apply table, copied verbatim (per-case body) from the
// old `switch g.phase` in gestPointerMove. A phase with no entry does nothing, matching the
// old switch's implicit default (gestIdle, gestPending, gestWiring all fell through to
// nothing).
var applyAction = map[gesturePhase]func(md *MoveDispatch, g *gestureState, ev rawInputMsg, tr *T.Trace){
	gestDragging: func(md *MoveDispatch, g *gestureState, ev rawInputMsg, tr *T.Trace) {
		if md.applyNodeDragTarget(ev) {
			g.prevX, g.prevY = ev.X, ev.Y
		}
	},
	gestRotating: func(md *MoveDispatch, g *gestureState, ev rawInputMsg, tr *T.Trace) {
		g.smoothX += rotSmoothAlpha * (ev.X - g.smoothX)
		g.smoothY += rotSmoothAlpha * (ev.Y - g.smoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.smoothX, g.smoothY
		md.applyOrbit(smoothEv, tr)
		g.prevX, g.prevY = g.smoothX, g.smoothY
	},
	gestHandhold: func(md *MoveDispatch, g *gestureState, ev rawInputMsg, tr *T.Trace) {
		g.smoothX += rotSmoothAlpha * (ev.X - g.smoothX)
		g.smoothY += rotSmoothAlpha * (ev.Y - g.smoothY)
		smoothEv := ev
		smoothEv.X, smoothEv.Y = g.smoothX, g.smoothY
		md.applyOrbitLocked(smoothEv, tr)
		g.prevX, g.prevY = g.smoothX, g.smoothY
	},
	gestPortMove: func(md *MoveDispatch, g *gestureState, ev rawInputMsg, tr *T.Trace) {
		md.applyPortMove(ev)
		g.prevX, g.prevY = ev.X, ev.Y
	},
}
