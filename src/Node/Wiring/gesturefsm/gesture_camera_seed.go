package gesturefsm

import (
	"math"

	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Node/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/spatial"
)

func (g *GestureState) BeginSphereRotation(vp Camera.Viewpoint, heldCenters func() map[string]spatial.Vec3, ev inputcodec.RawInputMsg) {
	pivot := Camera.FocusAhead(vp, heldCenters())
	g.RotPivot = pivot

	eye := Camera.EyeOf(vp)
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	ndcX, ndcY, _ := Camera.ProjectNDC(pivot, eye, basis, g.Fov, g.Rect.Aspect())
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
