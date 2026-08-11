package gesturefsm

// viewpoint_state.go — ViewpointState owns the polar camera viewpoint value plus its
// set/orbit/zoom/pan mutators. Lifted out of package Wiring alongside GestureState (same
// task, see gesture_state.go's package doc comment) — it has no dependency on *uiState/
// *MoveDispatch/*moverRegistry/*layoutQuantizer, only on geom and Trace. MoveDispatch
// (nodes/Wiring/viewpoint_state.go) keeps its own thin delegating methods
// (SetViewpoint/Viewpoint/EmitViewpoint/PanViewpoint) and the VIEW-frame-emitting callers
// (applyOrbit/applyOrbitLocked/gestHome/gestWheel) — those stay in package Wiring because
// emitting the VIEW frame needs md.sw/md.RT, per
// docs/planning/movedispatch-decomposition.md's write-then-emit split.
//
// The viewpoint value is embedded so callers that reach through the field (e.g. tests
// asserting md.ui.vp.LockedAxis) keep resolving, and the *geom.Viewpoint navigation ops
// (Orbit/OrbitLocked/Zoom/Pan) promote onto ViewpointState.

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// ViewpointState carries the camera viewpoint and its emit/navigation methods.
type ViewpointState struct {
	geom.Viewpoint
	// Persist, when non-nil, is called with the current viewpoint after every EmitViewpoint
	// so a gesture-driven change is persisted to camera.json. nil until armed by
	// MoveDispatch.EnableViewpointPersist (after the startup seed), so the seed's own emit
	// does not write. Owned by MoveDispatch; the debounce/write live in the persister.
	Persist func(geom.Viewpoint)
}

// SetViewpoint installs a known camera state without emitting. Used by the "set"
// viewpoint op to seed the viewpoint from persisted or initial values, followed by
// EmitViewpoint to broadcast it. Also clears any locked rotation axis from a prior
// handhold gesture so the next gesture starts fresh.
func (v *ViewpointState) SetViewpoint(pivot wire.Vec3, r float64, pos, up geom.Dir) {
	v.Pivot = pivot
	v.R = r
	v.Pos = pos
	v.Up = up
	v.LockedAxis = nil
}

// EmitViewpoint persists the current camera viewpoint state (the VIEW frame carrying it
// is written by the caller, MoveDispatch.EmitViewpoint in package Wiring).
func (v *ViewpointState) EmitViewpoint(tr *T.Trace) {
	// Persist the just-emitted viewpoint (debounced, off the hot path) when armed —
	// independent of the trace sink. Every gesture viewpoint change (orbit/zoom/pan/home)
	// flows through EmitViewpoint, so this is the single chokepoint for the write side.
	if v.Persist != nil {
		v.Persist(v.Viewpoint)
	}
}

// OrbitViewpoint applies a great-circle orbit (carrying from→to) and runs the persist hook.
// The caller (MoveDispatch.OrbitViewpoint, package Wiring) emits the VIEW frame.
func (v *ViewpointState) OrbitViewpoint(from, to geom.Dir, tr *T.Trace) {
	v.Orbit(from, to)
	v.EmitViewpoint(tr)
}

// OrbitLockedViewpoint applies a handhold-constrained orbit: the first call locks the
// rotation axis from the from→to arc; subsequent calls keep the same axis. The lock is
// cleared by the next SetViewpoint. Runs the persist hook; the caller emits the frame.
func (v *ViewpointState) OrbitLockedViewpoint(from, to geom.Dir, tr *T.Trace) {
	v.OrbitLocked(from, to)
	v.EmitViewpoint(tr)
}

// ZoomViewpoint scales the orbit radius by factor and runs the persist hook; the caller
// emits the frame.
func (v *ViewpointState) ZoomViewpoint(factor float64, tr *T.Trace) {
	v.Zoom(factor)
	v.EmitViewpoint(tr)
}

// PanViewpoint slides the orbit pivot by a world delta and runs the persist hook; the
// caller emits the frame.
func (v *ViewpointState) PanViewpoint(delta wire.Vec3, tr *T.Trace) {
	v.Pan(delta)
	v.EmitViewpoint(tr)
}
