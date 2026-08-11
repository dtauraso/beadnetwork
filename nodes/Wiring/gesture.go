package Wiring

// gesture.go — the GESTURE STATE MACHINE's OWNED STATE. It consumes RAW pointer/wheel input
// (forwarded fire-and-forget from TS behind USE_RAW_INPUT) plus the stateless raycast hit,
// and owns the in-progress gesture bookkeeping (origin, button, phase, frozen rotation
// frame) that decides what the raw input MEANS — orbit / zoom / pan / drag / wire. This is
// the one place gesture state lives (the spec's "gesture state machine lives in Go, in one
// place"); TS holds none of it. This file keeps only the FSM's TYPES and STATE (phase enum,
// gestureState, gestureRect, pixelToNDC, reset). The entry point and dispatch table live in
// gesture_dispatch.go; the per-event phase handlers that own the transitions live in
// gesture_handlers.go; the pointer-down hit classification lives in gesture_hitclassify.go;
// the leaf ACTIONS invoked by the phase handlers (orbit/drag/hover/select) live in
// gesture_actions.go; hit-resolution helpers (row index → topology identity) live in
// gesture_hit.go.
//
// The camera OUTCOMES are produced through the already-tested polar viewpoint ops
// (OrbitViewpoint / ZoomViewpoint / PanViewpoint → spherical.go), fed by the renderer-edge
// camera math in gesture_camera.go (ported formula-for-formula from the TS handlers). This
// file adds no new orbit/rotation math — it only sequences gestures and calls the ported
// helpers.
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
// Phase 7 closed the interaction gaps: click-select is Go-owned (md.ui.sel.Selected +
// KindSelect trace → buffer Selected column); handhold-constrained orbit is ported here
// formula-faithfully from interaction-handlers.ts. gestWiring and gestPortMove (an
// unconnected/connected port drag) were removed with port geometry
// (docs/bead-model/channels-not-ports.md): a port is a load-time channel-binding ROLE only, never
// drawn or hit-testable, so the "port" raycast-hit kind that fed both phases can never
// fire. They were already dead in effect before this deletion — wire-drop created no edge
// (the create/delete edit ops were removed end-to-end) and port-move only snapped a
// ring-anchor index that no longer exists.
//
// The FSM's actual state (the phase enum, GestureState, GestureRect, and their state-only
// methods) now lives in nodes/Wiring/gesturefsm — lifted out per
// docs/planning/movedispatch-decomposition.md's "6." section. Every call site in this package
// names gesturefsm.GestureState/gesturefsm.GestureRect/gesturefsm.GesturePhase/
// gesturefsm.GestIdle etc. directly — no alias shim. The entry points, dispatch tables, and
// leaf actions that read/write *uiState stay here because uiState itself could not lift
// (item 5: view_stream.go reads md.ui.ov/md.ui.vp/md.ui.sceneSphere by unexported field).
