package dispatch

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
)

// gesture_helpers_test.go — shared fixtures for driving the gesture state machine
// (gesture.go) with raw pointer/wheel sequences and asserting the FSM state transitions +
// camera OUTCOMES (viewpoint pose changes). Uses a MoveDispatch with no registered node
// geometry (empty heldCenters → deterministic region-focus fallback), so the outcomes are
// hand-computable. mr is a real, empty moverreg.MoverRegistry (moverreg.New()) rather than a
// zero value: MoverRegistry's fields are unexported outside its own package, so a caller
// that later writes into md.mr.NodeGeoms() (e.g. TestGesturePressReleaseNoMoveSelects) needs
// the map already initialized.
func newGestureMD(v geom.Viewpoint) *MoveDispatch {
	md := &MoveDispatch{mr: moverreg.New()}
	md.UI.VP.Viewpoint = v
	return md
}

// canonical "+Z camera looking at origin, up +Y, r=100" viewpoint.
func canonicalViewpoint() geom.Viewpoint {
	return geom.Viewpoint{Pivot: vec3{X: 0, Y: 0, Z: 0}, R: 100, Pos: geom.Dir{Theta: math.Pi / 2, Phi: math.Pi / 2}, Up: geom.Dir{Theta: 0, Phi: 0}}
}

func rawEvent(kind string, x, y float64) inputcodec.RawInputMsg {
	return inputcodec.RawInputMsg{
		Kind: kind, X: x, Y: y,
		RectLeft: 0, RectTop: 0, RectWidth: 800, RectHeight: 600,
		Button: 0, Fov: 50,
		Hit: inputcodec.RawHit{Kind: "empty"},
	}
}
