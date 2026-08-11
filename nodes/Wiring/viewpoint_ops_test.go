package Wiring

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// viewpoint_ops_test.go — the Zoom/Pan/Orbit viewpoint ops mutate viewpointState's own
// fields correctly; the underlying orbit/zoom/pan math is verified in spherical_test.go /
// viewpoint tests, these assert the op wiring only. The RowEvent/VIEW-frame side of this
// (Decentralized, Step C, memory/feedback_no_single_writer_bridge.md) is a MoveDispatch-level concern,
// covered by TestMoveDispatchViewpointDelegatorsEmit below and viewpoint_bridge_test.go.

// TestZoomViewpointEmitsRadius: ZoomViewpoint scales r.
func TestZoomViewpointEmitsRadius(t *testing.T) {
	tr := T.New()
	vp := &gesturefsm.ViewpointState{}
	vp.SetViewpoint(vec3{}, 100, geom.Dir{Theta: 1.0}, geom.Dir{Theta: 1.5708})

	vp.ZoomViewpoint(0.5, tr)
	if vp.R != 50 {
		t.Fatalf("Zoom(0.5) on r=100: r=%v, want 50", vp.R)
	}
}

// TestPanViewpointEmitsPivot: PanViewpoint slides the pivot.
func TestPanViewpointEmitsPivot(t *testing.T) {
	tr := T.New()
	vp := &gesturefsm.ViewpointState{}
	vp.SetViewpoint(vec3{X: 1, Y: 2, Z: 3}, 100, geom.Dir{Theta: 1.0}, geom.Dir{Theta: 1.5708})

	vp.PanViewpoint(vec3{X: 10, Y: 0, Z: -3}, tr)
	if vp.Pivot.X != 11 || vp.Pivot.Y != 2 || vp.Pivot.Z != 0 {
		t.Fatalf("Pan: pivot=(%v,%v,%v), want (11,2,0)", vp.Pivot.X, vp.Pivot.Y, vp.Pivot.Z)
	}
}

// TestOrbitViewpointEmitsMovedPos: OrbitViewpoint carries pos from→to and changes the
// pos direction.
func TestOrbitViewpointEmitsMovedPos(t *testing.T) {
	tr := T.New()
	vp := &gesturefsm.ViewpointState{}
	before := geom.Dir{Theta: 1.0, Phi: 0.0}
	vp.SetViewpoint(vec3{}, 100, before, geom.Dir{Theta: 1.5708})

	vp.OrbitViewpoint(geom.Dir{Theta: 1.0, Phi: 0.0}, geom.Dir{Theta: 1.2, Phi: 0.3}, tr)
	if vp.Pos.Theta == before.Theta && vp.Pos.Phi == before.Phi {
		t.Fatalf("Orbit did not change pos: still (%v,%v)", vp.Pos.Theta, vp.Pos.Phi)
	}
}

// TestMoveDispatchViewpointDelegatorsEmit: the MoveDispatch delegators (Zoom/Pan/Orbit)
// forward to md.ui.vp; the view-owner goroutine (the real callers: gesture_actions.go,
// gesture_handlers.go) emits the camera RowEvent onto the VIEW stream after each mutation —
// this test drives that same mutate-then-emit shape at the call site, per
// docs/planning/movedispatch-decomposition.md's write-then-emit split.
func TestMoveDispatchViewpointDelegatorsEmit(t *testing.T) {
	tr := T.New()
	md := &MoveDispatch{}
	var events []wire.RowEvent
	captureViewFrameKinds(md, &events)
	md.SetViewpoint(vec3{}, 100, geom.Dir{Theta: 1.0}, geom.Dir{Theta: 1.5708})

	md.ui.ZoomViewpoint(0.5, tr)
	md.emitViewFrame(cameraViewEvent())
	md.PanViewpoint(vec3{X: 5}, tr)
	md.emitViewFrame(cameraViewEvent())
	md.ui.OrbitViewpoint(geom.Dir{Theta: 1.0, Phi: 0.0}, geom.Dir{Theta: 1.1, Phi: 0.1}, tr)
	md.emitViewFrame(cameraViewEvent())

	if n := countCameraEvents(events); n < 3 {
		t.Fatalf("expected >=3 camera events from delegators, got %d", n)
	}
	// Confirm the final viewpoint state reflects the zoom+pan (r halved, pivot moved).
	if md.ui.vp.R != 50 || md.ui.vp.Pivot.X != 5 {
		t.Fatalf("delegator state: r=%v pivot.X=%v, want r=50 pivot.X=5", md.ui.vp.R, md.ui.vp.Pivot.X)
	}
}
