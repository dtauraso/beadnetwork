// Package gesturefsm holds the gesture FSM's OWNED STATE: the phase enum, GestureState,
// GestureRect, and their pure state-only methods (PixelToNDC, Reset, Aspect) — lifted out of
// package Wiring per docs/planning/movedispatch-decomposition.md's "6." section and
// docs/planning/gesture-actor.md's step 3.
//
// This package holds ONLY the state and its own methods — not the FSM's entry points
// (gestPointerDown/Move/Up, HandleRawInput, gestHome, gestWheel) or the leaf actions that
// read/write *uiState (beginSphereRotation, applyNodeDragTarget, commitDragStart, setHover).
// Those stay on MoveDispatch/in package Wiring: they are reached by view_stream.go and other
// Wiring files through the unexported *uiState type directly, by field, so they cannot move
// without either dragging uiState (blocked — see movedispatch-decomposition.md item 5) or
// exporting uiState's entire private surface, which the task that authored this package
// declined to do.
package gesturefsm

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// GesturePhase is the FSM's current phase.
type GesturePhase int

const (
	GestIdle GesturePhase = iota
	GestPending
	GestRotating
	GestDragging
	GestHandhold
)

// GestureState is the FSM's owned bookkeeping. Zero value = idle.
type GestureState struct {
	Phase GesturePhase

	// pointer-down snapshot + running previous position (client pixels)
	DownX, DownY float64
	PrevX, PrevY float64
	Button       int

	// SmoothX/SmoothY are the AVERAGING ("fat") cursor driving rotation: an exponential
	// moving average of the raw pointer position, so the rotation follows a continuously
	// blurred cursor (never holds, never freezes — just lags-free-smooths jitter). Seeded to
	// the raw position when a rotation drag (GestRotating/GestHandhold) begins.
	SmoothX, SmoothY float64
	// Secondary is true when the pointer-down was a SECONDARY (button 2) press — a
	// two-finger trackpad tap. Such a press is always a tap-select and NEVER converts to a
	// drag/rotate, so it stays GestPending through any finger drift and resolves to a
	// select on pointer-up.
	Secondary bool

	// empty-space rotation gate + the entity grabbed at pointer-down
	EmptyDown bool

	// node-drag target
	DragNode        string
	DragStartCenter wire.Vec3
	// DragGrabOffset is DragStartCenter minus the plane-hit at the drag-start commit
	// event — the vector from where the pointer actually grabbed the node to the
	// node's center, captured ONCE at drag start (commitDragStart) and reapplied on
	// every subsequent move (applyNodeDragTarget). Per-drag state belongs at the FSM
	// drag-start edge, not inside the move path: computing this offset per-move would
	// measure it against the node's ALREADY-MOVED center each event and cancel itself
	// out (memory/project_rootmove_is_per_pointer_move.md — RootMove runs on every
	// pointer-move, not once per drag; two prior bugs came from forgetting that).
	// Zero (its default) degrades to today's centre-on-cursor behavior, which is the
	// right fallback when the drag-start ray is parallel to the drag plane.
	DragGrabOffset wire.Vec3

	// handhold-constrained orbit gate (set at pointer-down on a handhold hit).
	HandholdDown bool

	// rotation frame, FROZEN at gesture start (mirrors beginSphereRotation): the pivot,
	// its screen-pixel center, and pixels-per-radian for ScreenToPolar.
	RotPivot     wire.Vec3
	RotCx, RotCy float64
	RotPxPerRad  float64

	// per-gesture render params captured from the raw events
	Fov  float64
	Rect GestureRect
}

// GestureRect is the last render viewport (client-pixel rect) a gesture's raw input
// reported.
type GestureRect struct{ Left, Top, Width, Height float64 }

// Aspect returns the rect's width/height, defaulting to 1 for a zero-height rect.
func (r GestureRect) Aspect() float64 {
	if r.Height == 0 {
		return 1
	}
	return r.Width / r.Height
}

// PixelToNDC mirrors geometry-helpers.ts pixelToNDC.
func (g *GestureState) PixelToNDC(x, y float64) (nx, ny float64) {
	nx = ((x-g.Rect.Left)/g.Rect.Width)*2 - 1
	ny = -((y-g.Rect.Top)/g.Rect.Height)*2 + 1
	return nx, ny
}

// Reset clears the gesture FSM back to idle at the end of every gesture (pointer-up).
// It also clears vp.LockedAxis (the handhold-constrained-orbit rotation axis frozen at
// gesture start — see geom.Viewpoint.LockedAxis's doc comment) so that field's own "nil
// between gestures" doc is actually true: lockedAxis is gesture-scoped state, exactly
// like DragNode above, it just happens to live on viewpoint instead of GestureState
// (frozen once per handhold gesture in orbit's lazy-init path). Today it is always
// overwritten before use anyway (every new gesture reseeds it via SetViewpoint/
// seedOrbitPivot before orbit ever reads it), so this had no live bug — but Reset() is the
// obvious single home for "gesture-scoped state ends here", so it belongs here rather than
// living only as an unenforced comment.
func (g *GestureState) Reset(vp *geom.Viewpoint) {
	g.Phase = GestIdle
	g.EmptyDown = false
	g.DragNode = ""
	g.DragGrabOffset = wire.Vec3{}
	g.HandholdDown = false
	g.Secondary = false
	vp.LockedAxis = nil
}
