// Package gesture is the GESTURE STATE MACHINE. It consumes RAW pointer/wheel input
// (forwarded fire-and-forget from TS behind USE_RAW_INPUT) plus the stateless raycast hit,
// and owns the in-progress gesture bookkeeping (origin, button, phase, frozen rotation
// frame) that decides what the raw input MEANS — orbit / zoom / pan / drag / wire. This is
// the one place gesture state lives (the spec's "gesture state machine lives in Go, in one
// place"); TS holds none of it. This file keeps only the package doc; the FSM's owned
// STATE TYPES (phase enum, GestureState, GestureRect) live in nodes/Wiring/gesturefsm. The
// entry point and dispatch table live in gesture_dispatch.go; the per-event phase handlers
// that own the transitions live in gesture_handlers.go; the pointer-down hit classification
// lives in gesture_hitclassify.go; the leaf ACTIONS invoked by the phase handlers
// (orbit/drag/hover/select) live in gesture_actions.go; the commit/apply tables live in
// gesture_graph.go.
//
// The camera OUTCOMES are produced through the already-tested polar viewpoint ops
// (OrbitViewpoint / ZoomViewpoint / PanViewpoint → viewstate/gesturefsm), fed by the
// renderer-edge camera math in nodes/Wiring/geom (ported formula-for-formula from the TS
// handlers). This package adds no new orbit/rotation math — it only sequences gestures and
// calls the ported helpers.
//
// States:
//
//	idle      — nothing in progress.
//	pending   — pointer is down; no movement yet. Resolves to a drag/rotate on the first
//	            pixel of actual displacement from the press point, or to a click/wire on
//	            pointer-up with no movement at all.
//	rotating  — empty-space great-circle orbit about a frozen region-focus pivot.
//	dragging  — node body drag (world target on a camera-facing plane → RootMove).
//	handhold  — a handhold grab-sphere is dragged for axis-locked (constrained) orbit.
//
// This package was lifted out of package dispatch (docs/planning/movedispatch-decomposition.md
// §31): every function here used to be a *dispatch.MoveDispatch method or take one as a
// parameter, but the only field access that ever reached an unexported MoveDispatch field
// was md.ctx (a context.Context) — now threaded as an explicit parameter (Deps.Ctx), the
// same shape ApplyEdit→applyUpdate used for the stdin cluster in §30. Every other field
// this package reads (MR, UI, LQ, RT) was already an exported sub-object on MoveDispatch,
// so no alias shim or interface was needed — this package names moverreg.MoverRegistry,
// viewstate.UIState, layoutquant.LayoutQuantizer, and rowtables.RowTables directly.
//
// dispatch.MoveDispatch.HandleRawInput is now a two-line delegator into this package's
// HandleRawInput, bundling its own MR/UI/LQ/RT/ctx into a Deps value. This creates one new
// import edge (dispatch → gesture, a caller depending on what it calls) and closes off the
// reverse: this package must never import dispatch.
package gesture
