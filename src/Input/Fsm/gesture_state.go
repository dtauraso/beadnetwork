package Fsm

import (
	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/spatial"
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

	DragNode        string
	DragStartCenter spatial.Vec3

	DragGrabOffset spatial.Vec3

	HandholdDown bool

	RotPivot     spatial.Vec3
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
	g.DragNode = ""
	g.DragGrabOffset = spatial.Vec3{}
	g.HandholdDown = false
	g.Secondary = false
	vp.LockedAxis = nil
}
