package dispatch

import (
	"math"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesturefsm"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// gesture_camera_outcomes_test.go — drag/wheel/handhold gestures and the camera pose
// (viewpoint) outcomes they produce.

// Empty-space drag orbits the camera about a frozen region-focus pivot: pivot + radius are
// preserved (rigid orbit) while pos changes.
func TestGestureEmptyDragOrbits(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())

	down := rawEvent("pointerdown", 400, 300)
	md.HandleRawInput(down, nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestPending || !md.UI.Gest.EmptyDown {
		t.Fatalf("after pointerdown: phase=%v emptyDown=%v", md.UI.Gest.Phase, md.UI.Gest.EmptyDown)
	}
	// Orbit pivot is the content ahead (focusAhead). Empty centers → a point on the view axis a
	// fixed distance ahead: eye=(0,0,100), forward=(0,0,-1), focusMin=10 → (0,0,90).
	if !vecClose(md.UI.Gest.RotPivot, vec3{X: 0, Y: 0, Z: 90}, 1e-9) {
		t.Fatalf("rotPivot=%v want focus-ahead (0,0,90)", md.UI.Gest.RotPivot)
	}

	// First move past the slop: transitions to rotating and seeds the viewpoint. The first
	// frame's arc is ~zero (prev==curr), so pose is essentially the seeded one.
	md.HandleRawInput(rawEvent("pointermove", 420, 300), nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestRotating {
		t.Fatalf("after slop-cross move: phase=%v want rotating", md.UI.Gest.Phase)
	}
	if !vecClose(md.UI.VP.Pivot, vec3{X: 0, Y: 0, Z: 90}, 1e-9) {
		t.Fatalf("seed pivot=%v want focus-ahead (0,0,90)", md.UI.VP.Pivot)
	}
	if math.Abs(md.UI.VP.R-10) > 1e-9 {
		t.Fatalf("seed r=%v want 10 (eye→focus-ahead)", md.UI.VP.R)
	}
	posBefore := md.UI.VP.Pos
	rBefore, pivotBefore := md.UI.VP.R, md.UI.VP.Pivot

	// Second move: genuine cursor delta → orbit. pos must change; r + pivot preserved.
	md.HandleRawInput(rawEvent("pointermove", 480, 320), nil, nil)
	if math.Abs(md.UI.VP.R-rBefore) > 1e-9 {
		t.Fatalf("orbit changed r: %v != %v", md.UI.VP.R, rBefore)
	}
	if !vecClose(md.UI.VP.Pivot, pivotBefore, 1e-9) {
		t.Fatalf("orbit moved pivot: %v != %v", md.UI.VP.Pivot, pivotBefore)
	}
	if geom.AngularDistance(md.UI.VP.Pos, posBefore) < 1e-6 {
		t.Fatalf("orbit did not change pos (dir stayed %v)", md.UI.VP.Pos)
	}

	md.HandleRawInput(rawEvent("pointerup", 480, 320), nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestIdle {
		t.Fatalf("after pointerup: phase=%v want idle", md.UI.Gest.Phase)
	}
}

// Plain wheel pans the pivot (screen-space slide); ctrl+wheel dollies (pivot translation
// toward the cursor target). Both leave the radius set by the region-focus seed.
func TestGestureWheelStrafesCamera(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	pivotBefore := md.UI.VP.Pivot
	centerBefore := md.UI.SceneSphere.Center
	ev := rawEvent("wheel", 400, 300)
	ev.DeltaX = 10
	ev.DeltaY = 0
	md.HandleRawInput(ev, nil, nil)
	// Lateral pan strafes the CAMERA (free-camera model); the fixed scene does not move.
	if vecClose(md.UI.VP.Pivot, pivotBefore, 1e-9) {
		t.Fatalf("plain wheel did not strafe the camera (pivot stayed %v)", md.UI.VP.Pivot)
	}
	if !vecClose(md.UI.SceneSphere.Center, centerBefore, 1e-9) {
		t.Fatalf("plain wheel moved the scene center %v; the scene must stay fixed", md.UI.SceneSphere.Center)
	}
}

// Plain-wheel PAN must fire regardless of what the raycast hit is under the cursor: the
// gesture FSM's wheel path is hit-independent for a plain (non-ctrl) wheel. This pins that a
// node/edge hit does NOT suppress or divert the pan (the TS-side validator drop of "edge"
// hits was the real bug; this guards the Go contract the fix relies on).
func TestGestureWheelPansOverNodeAndEdgeHit(t *testing.T) {
	for _, h := range []inputcodec.RawHit{
		{Kind: "node", NodeRow: 0},
		{Kind: "edge", EdgeRow: 0},
		{Kind: "port", PortRow: 0},
	} {
		md := newGestureMD(canonicalViewpoint())
		md.RT.NodeRowTable = []string{"N7"}
		before := md.UI.VP.Pivot
		ev := rawEvent("wheel", 400, 300)
		ev.DeltaX = 10
		ev.DeltaY = 0
		ev.Hit = h
		md.HandleRawInput(ev, nil, nil)
		if vecClose(md.UI.VP.Pivot, before, 1e-9) {
			t.Fatalf("plain wheel with %s hit did not strafe the camera (pivot stayed %v)", h.Kind, md.UI.VP.Pivot)
		}
	}
}

func TestGestureCtrlWheelZoomsToCursor(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	ev := rawEvent("wheel", 400, 300)
	ev.Ctrl = true
	ev.DeltaY = 1
	md.HandleRawInput(ev, nil, nil)
	// ctrl-wheel dollies the camera along the cursor→node ray KEEPING orientation (node stays
	// under the mouse — no re-aim). Empty centers → target=regionFocus=(0,0,90), eye=(0,0,100),
	// rayDir=(0,0,-1), distP=10. The fractional step (10*(1-1.01)=-0.1) is below the pass-through
	// floor (minStep = vp.r*(zoomBase-1) = 1), so the camera moves minStep AWAY (DeltaY=1) →
	// pivot.Z = +1.
	wantZ := 100 * (geom.GestureZoomBase - 1)
	if math.Abs(md.UI.VP.Pivot.Z-wantZ) > 1e-9 || math.Abs(md.UI.VP.Pivot.X) > 1e-9 {
		t.Fatalf("ctrl-wheel pivot=%v want Z≈%v (dolly toward cursor)", md.UI.VP.Pivot, wantZ)
	}
	// The look direction is unchanged (no re-aim).
	if geom.AngularDistance(md.UI.VP.Pos, canonicalViewpoint().Pos) > 1e-9 {
		t.Fatalf("ctrl-wheel re-aimed the camera (pos changed); zoom-to-cursor must keep orientation")
	}
}

// A handhold grab resolves (past the slop) to axis-locked orbit: the camera pose changes
// (pos moves) while the pivot + radius stay fixed, just like a free orbit.
func TestGestureHandholdOrbits(t *testing.T) {
	md := newGestureMD(canonicalViewpoint())
	down := rawEvent("pointerdown", 400, 300)
	down.Hit = inputcodec.RawHit{Kind: "handhold"}
	md.HandleRawInput(down, nil, nil)
	if !md.UI.Gest.HandholdDown || md.UI.Gest.Phase != gesturefsm.GestPending {
		t.Fatalf("after handhold down: handholdDown=%v phase=%v", md.UI.Gest.HandholdDown, md.UI.Gest.Phase)
	}
	md.HandleRawInput(rawEvent("pointermove", 460, 320), nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestHandhold {
		t.Fatalf("phase=%v want handhold", md.UI.Gest.Phase)
	}
	rBefore, pivotBefore, posBefore := md.UI.VP.R, md.UI.VP.Pivot, md.UI.VP.Pos
	md.HandleRawInput(rawEvent("pointermove", 520, 360), nil, nil)
	if math.Abs(md.UI.VP.R-rBefore) > 1e-9 {
		t.Fatalf("handhold orbit changed r: %v != %v", md.UI.VP.R, rBefore)
	}
	if !vecClose(md.UI.VP.Pivot, pivotBefore, 1e-9) {
		t.Fatalf("handhold orbit moved pivot: %v != %v", md.UI.VP.Pivot, pivotBefore)
	}
	if geom.AngularDistance(md.UI.VP.Pos, posBefore) < 1e-6 {
		t.Fatalf("handhold orbit did not change pos (stayed %v)", md.UI.VP.Pos)
	}
	md.HandleRawInput(rawEvent("pointerup", 520, 360), nil, nil)
	if md.UI.Gest.Phase != gesturefsm.GestIdle {
		t.Fatalf("after handhold up phase=%v want idle", md.UI.Gest.Phase)
	}
}
