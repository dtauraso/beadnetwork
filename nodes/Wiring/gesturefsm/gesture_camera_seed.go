package gesturefsm

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func (g *GestureState) BeginSphereRotation(vp camera.Viewpoint, heldCenters func() map[string]spatial.Vec3, ev inputcodec.RawInputMsg) {
	pivot := camera.FocusAhead(vp, heldCenters())
	g.RotPivot = pivot

	eye := camera.EyeOf(vp)
	basis := camera.BasisFromViewpoint(vp.Pos, vp.Up)
	ndcX, ndcY, _ := camera.ProjectNDC(pivot, eye, basis, g.Fov, g.Rect.Aspect())
	g.RotCx = ((ndcX+1)/2)*g.Rect.Width + g.Rect.Left
	g.RotCy = ((-ndcY+1)/2)*g.Rect.Height + g.Rect.Top

	_, csRadius := polar.ContentSphereOf(heldCenters())
	pivotDist := eye.Sub(pivot).Length()
	fovRad := g.Fov * math.Pi / 180
	rpx := (g.Rect.Height / 2) / math.Tan(fovRad/2)
	if pivotDist > 0 {
		rpx *= csRadius / pivotDist
	}
	g.RotPxPerRad = math.Max(rpx*(2/math.Pi), 1)
}
