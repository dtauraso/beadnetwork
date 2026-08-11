package dispatch

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
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// SetViewpoint/EmitViewpoint delegators were deleted: md.UI is now exported
// (nodes/Wiring/viewstate.UIState), so runtopology (SetViewpoint's/EmitViewpoint's only
// out-of-package callers) reaches md.UI.VP.SetViewpoint/md.UI.VP.EmitViewpoint directly —
// see docs/planning/gesture-actor.md's payoff commit.

// cameraViewEvent is the single Camera event every camera-changing delegator below hands
// to emitViewFrame. Camera decodes entirely from the VIEW frame's own Camera block (see
// buffer-log.ts's decodeEventLine "camera" case) — no row identity to resolve.
func cameraViewEvent() []wire.RowEvent {
	return []wire.RowEvent{{Kind: T.KindCamera, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}}
}

// OrbitViewpoint/OrbitLockedViewpoint/ZoomViewpoint moved onto viewstate.UIState itself
// (docs/planning/gesture-actor.md's lift) — they are pure promotions onto the owned VP with
// no dependency on any Wiring-only type, so they lift cleanly; Wiring's gesture_actions.go
// now calls md.UI.OrbitViewpoint(...)/md.UI.OrbitLockedViewpoint(...) directly.

// PanViewpoint: a dolly is a pure CAMERA move (the eye translates toward the cursor). It
// must NOT move the scene sphere: coupling them left md.ui.sceneSphere.Center diverged from
// the movers' held center until a later broadcast reconciled it with a jump (the "zoom got
// canceled" symptom). Nothing moves the sphere — MODEL.md: "It is established once and
// never moves." Pan-moves-the-sphere is REJECTED doctrine, not a gap to fill; if it is ever
// revisited it must be its own gesture, never a side effect of a camera move. Callers reach
// this through md.UI.VP.PanViewpoint directly (gesturefsm.ViewpointState's own method).
