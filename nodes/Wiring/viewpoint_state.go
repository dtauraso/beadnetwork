package Wiring

// viewpoint_state.go — viewpointState owns the polar camera viewpoint value plus its
// set/orbit/zoom/pan mutators and the camera-trace emit. It is owned as a field by
// MoveDispatch (md.ui.vp), which exposes thin delegating methods; extracting it here keeps
// move_dispatch.go focused on the dispatch registry. There is no goroutine — callers
// serialize externally (the stdin reader runs in a single goroutine).
//
// viewpointState itself now lives in nodes/Wiring/gesturefsm (ViewpointState) — lifted out
// per docs/planning/movedispatch-decomposition.md's "6." section, since it has no
// dependency on *uiState/*MoveDispatch/*moverRegistry/*layoutQuantizer. md.ui.vp is typed
// gesturefsm.ViewpointState directly (no alias shim), and the *geom.Viewpoint navigation ops
// (Orbit/OrbitLocked/Zoom/Pan) still promote onto it. The delegators below (which need
// md.sw/md.RT to emit the VIEW frame, per the write-then-emit split) stay here.

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// Camera viewpoint API — thin delegators to the owned viewpointState above. SetViewpoint
// and Viewpoint both have out-of-package callers (runtopology passes md.SetViewpoint as a
// func value to scenecamera.SeedInitialViewpoint; nodes/Wiring/scenecamera's own tests read
// md.Viewpoint()) that cannot reach the unexported md.ui.vp field directly, so neither can
// be deleted without exporting ui/vp.

func (md *MoveDispatch) SetViewpoint(pivot vec3, r float64, pos, up geom.Dir) {
	md.UI.VP.SetViewpoint(pivot, r, pos, up)
}

// Viewpoint returns the CURRENT camera viewpoint (pivot/r/pos/up/lockedAxis). Read-only
// accessor for callers outside this package — e.g. an external test of a package that
// itself takes *MoveDispatch, such as nodes/Wiring/scenecamera's own tests asserting what
// SeedInitialViewpoint installed.
func (md *MoveDispatch) Viewpoint() geom.Viewpoint {
	return md.UI.VP.Viewpoint
}

// cameraViewEvent is the single Camera event every camera-changing delegator below hands
// to emitViewFrame. Camera decodes entirely from the VIEW frame's own Camera block (see
// buffer-log.ts's decodeEventLine "camera" case) — no row identity to resolve.
func cameraViewEvent() []wire.RowEvent {
	return []wire.RowEvent{{Kind: T.KindCamera, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}}
}

// EmitViewpoint/OrbitViewpoint/OrbitLockedViewpoint/ZoomViewpoint/PanViewpoint mutate the
// owned viewpointState only. The VIEW frame is emitted by the caller — always the
// view-owner goroutine (RunStdinReader, via the gesture handlers) — not here, per
// docs/planning/movedispatch-decomposition.md's write-then-emit split.
func (md *MoveDispatch) EmitViewpoint(tr *T.Trace) {
	md.UI.VP.EmitViewpoint(tr)
}

// OrbitViewpoint/OrbitLockedViewpoint/ZoomViewpoint moved onto viewstate.UIState itself
// (docs/planning/gesture-actor.md's lift) — they are pure promotions onto the owned VP with
// no dependency on any Wiring-only type, so they lift cleanly; Wiring's gesture_actions.go
// now calls md.UI.OrbitViewpoint(...)/md.UI.OrbitLockedViewpoint(...) directly.

func (md *MoveDispatch) PanViewpoint(delta vec3, tr *T.Trace) {
	// A dolly is a pure CAMERA move (the eye translates toward the cursor). It must NOT move the
	// scene sphere: coupling them left md.ui.sceneSphere.Center diverged from the movers' held
	// center until a later broadcast reconciled it with a jump (the "zoom got canceled"
	// symptom). Nothing moves the sphere — MODEL.md: "It is established once and never moves."
	// Pan-moves-the-sphere is REJECTED doctrine, not a gap to fill; if it is ever revisited it
	// must be its own gesture, never a side effect of a camera move.
	md.UI.VP.PanViewpoint(delta, tr)
}
