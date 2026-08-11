package gesturefsm

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// BeginSphereRotation freezes the orbit pivot, its screen-pixel center, and pixels-per-radian
// for the whole gesture, writing only this GestureState's own fields. The pivot is the
// CONTENT DIRECTLY AHEAD (focusAhead): the node the camera is most pointed at, at its depth
// on the view-center ray. So rotate orbits whatever you have flown to and centered (fly to a
// node → rotate spins around it), the orbit depth tracks what you look at, and — because the
// pivot is on the view axis — it does not re-aim the camera.
//
// Lifted out of package Wiring (was the package-level beginSphereRotation in
// gesture_actions.go, which took *viewstate.UIState only to read vp/g and write g — every
// statement in its body was computation on locals/params plus writes to this type's own
// fields, so it moves as a method taking the viewpoint by value and the held-centers reader
// by closure, same treatment the Wiring call sites already used for mr/lq).
func (g *GestureState) BeginSphereRotation(vp geom.Viewpoint, heldCenters func() map[string]wire.Vec3, ev inputcodec.RawInputMsg) {
	pivot := geom.FocusAhead(vp, heldCenters())
	g.RotPivot = pivot

	eye := geom.EyeOf(vp)
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	ndcX, ndcY, _ := geom.ProjectNDC(pivot, eye, basis, ev.Fov, g.Rect.Aspect())
	g.RotCx = ((ndcX+1)/2)*g.Rect.Width + g.Rect.Left
	g.RotCy = ((-ndcY+1)/2)*g.Rect.Height + g.Rect.Top

	// Rotate sensitivity is ANCHORED TO THE ON-SCREEN CONTENT-SPHERE RADIUS: pixels-per-radian
	// scales by csRadius/pivotDist (the sphere's angular size), so a quarter-turn (pi/2) is
	// reached by dragging one on-screen content-sphere radius, at every zoom level. Without the
	// anchor, pi/2 required dragging nearly the full screen height and felt unreachable.
	_, csRadius := geom.ContentSphereOf(heldCenters())
	pivotDist := eye.Sub(pivot).Length()
	fovRad := ev.Fov * math.Pi / 180
	rpx := (g.Rect.Height / 2) / math.Tan(fovRad/2)
	if pivotDist > 0 {
		rpx *= csRadius / pivotDist
	}
	g.RotPxPerRad = math.Max(rpx*(2/math.Pi), 1)
}
