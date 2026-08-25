package Drag

import (
	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
)

type GesturePhase int

const (
	GestIdle GesturePhase = iota
	GestPending
	GestRotating
	GestDragging
	GestHandhold
)

type GestureState struct {
	Phase GesturePhase

	DownX, DownY float64
	PrevX, PrevY float64
	Button       int

	Secondary bool

	EmptyDown bool

	NodeDrag NodeDrag.Gesture

	HandholdDown bool

	RotPivot Vec3

	PressVP Camera.Viewpoint

	Fov  float64
	Rect GestureRect
}

type GestureRect struct{ Left, Top, Width, Height float64 }

func (r GestureRect) Aspect() float64 {
	if r.Height == 0 {
		return 1
	}
	return r.Width / r.Height
}

func (g *GestureState) Reset(vp *Camera.Viewpoint) {
	g.Phase = GestIdle
	g.EmptyDown = false
	g.NodeDrag.Clear()
	g.HandholdDown = false
	g.Secondary = false
}
