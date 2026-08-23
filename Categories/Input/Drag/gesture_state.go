package Drag

import (
	"github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
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

	SmoothX, SmoothY float64

	Secondary bool

	EmptyDown bool

	NodeDrag Node.DragGesture

	HandholdDown bool

	RotPivot     Vec3
	RotCx, RotCy float64
	RotPxPerRad  float64

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

func (g *GestureState) PixelToNDC(x, y float64) (nx, ny float64) {
	nx = ((x-g.Rect.Left)/g.Rect.Width)*2 - 1
	ny = -((y-g.Rect.Top)/g.Rect.Height)*2 + 1
	return nx, ny
}

func (g *GestureState) Reset(vp *Camera.Viewpoint) {
	g.Phase = GestIdle
	g.EmptyDown = false
	g.NodeDrag.Clear()
	g.HandholdDown = false
	g.Secondary = false
	vp.LockedAxis = nil
}
