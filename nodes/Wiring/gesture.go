package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
)

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
//   idle      — nothing in progress.
//   pending   — pointer is down; no movement yet. Resolves to a drag/rotate on the first
//               pixel of actual displacement from the press point, or to a click/wire on
//               pointer-up with no movement at all.
//   rotating  — empty-space great-circle orbit about a frozen region-focus pivot.
//   dragging  — node body drag (world target on a camera-facing plane → RootMove).
//   handhold  — a handhold grab-sphere is dragged for axis-locked (constrained) orbit.
//
// Phase 7 closed the interaction gaps: click-select is Go-owned (md.ui.sel.selected +
// KindSelect trace → buffer Selected column); handhold-constrained orbit is ported here
// formula-faithfully from interaction-handlers.ts. gestWiring and gestPortMove (an
// unconnected/connected port drag) were removed with port geometry
// (docs/bead-model/channels-not-ports.md): a port is a load-time channel-binding ROLE only, never
// drawn or hit-testable, so the "port" raycast-hit kind that fed both phases can never
// fire. They were already dead in effect before this deletion — wire-drop created no edge
// (the create/delete edit ops were removed end-to-end) and port-move only snapped a
// ring-anchor index that no longer exists.

type gesturePhase int

const (
	gestIdle gesturePhase = iota
	gestPending
	gestRotating
	gestDragging
	gestHandhold
)

// gestureState is the FSM's owned bookkeeping. Zero value = idle.
type gestureState struct {
	phase gesturePhase

	// pointer-down snapshot + running previous position (client pixels)
	downX, downY float64
	prevX, prevY float64
	button       int

	// smoothX/smoothY are the AVERAGING ("fat") cursor driving rotation: an exponential
	// moving average of the raw pointer position, so the rotation follows a continuously
	// blurred cursor (never holds, never freezes — just lags-free-smooths jitter). Seeded to
	// the raw position when a rotation drag (gestRotating/gestHandhold) begins.
	smoothX, smoothY float64
	// secondary is true when the pointer-down was a SECONDARY (button 2) press — a
	// two-finger trackpad tap. Such a press is always a tap-select and NEVER converts to a
	// drag/rotate, so it stays `gestPending` through any finger drift and resolves to a
	// select on pointer-up.
	secondary bool

	// empty-space rotation gate + the entity grabbed at pointer-down
	emptyDown bool

	// node-drag target
	dragNode        string
	dragStartCenter vec3
	// dragGrabOffset is dragStartCenter minus the plane-hit at the drag-start commit
	// event — the vector from where the pointer actually grabbed the node to the
	// node's center, captured ONCE at drag start (commitDragStart) and reapplied on
	// every subsequent move (applyNodeDragTarget). Per-drag state belongs at the FSM
	// drag-start edge, not inside the move path: computing this offset per-move would
	// measure it against the node's ALREADY-MOVED center each event and cancel itself
	// out (memory/project_rootmove_is_per_pointer_move.md — RootMove runs on every
	// pointer-move, not once per drag; two prior bugs came from forgetting that).
	// Zero (its default) degrades to today's centre-on-cursor behavior, which is the
	// right fallback when the drag-start ray is parallel to the drag plane.
	dragGrabOffset vec3

	// handhold-constrained orbit gate (set at pointer-down on a handhold hit).
	handholdDown bool

	// rotation frame, FROZEN at gesture start (mirrors beginSphereRotation): the pivot,
	// its screen-pixel center, and pixels-per-radian for screenToPolar.
	rotPivot     vec3
	rotCx, rotCy float64
	rotPxPerRad  float64

	// per-gesture render params captured from the raw events
	fov  float64
	rect gestureRect
}

type gestureRect struct{ left, top, width, height float64 }

func (r gestureRect) aspect() float64 {
	if r.height == 0 {
		return 1
	}
	return r.width / r.height
}

// pixelToNDC mirrors geometry-helpers.ts pixelToNDC.
func (g *gestureState) pixelToNDC(x, y float64) (nx, ny float64) {
	nx = ((x-g.rect.left)/g.rect.width)*2 - 1
	ny = -((y-g.rect.top)/g.rect.height)*2 + 1
	return nx, ny
}

// reset clears the gesture FSM back to idle at the end of every gesture (pointer-up).
// It also clears vp.LockedAxis (the handhold-constrained-orbit rotation axis frozen at
// gesture start — see geom.Viewpoint.LockedAxis's doc comment) so that field's own "nil
// between gestures" doc is actually true: lockedAxis is gesture-scoped state, exactly
// like dragNode above, it just happens to live on viewpoint
// instead of gestureState (frozen once per handhold gesture in orbit's lazy-init path).
// Today it is always overwritten before use anyway (every new gesture reseeds it via
// SetViewpoint/seedOrbitPivot before orbit ever reads it), so this had no live bug —
// but reset() is the obvious single home for "gesture-scoped state ends here", so it
// belongs here rather than living only as an unenforced comment.
func (g *gestureState) reset(vp *geom.Viewpoint) {
	g.phase = gestIdle
	g.emptyDown = false
	g.dragNode = ""
	g.dragGrabOffset = vec3{}
	g.handholdDown = false
	g.secondary = false
	vp.LockedAxis = nil
}
