package Drag

import (
	"math"

	"github.com/dtauraso/wirefold/Camera"
	"github.com/dtauraso/wirefold/Polar/polar"
)

func (g *GestureState) BeginSphereRotation(vp Camera.Viewpoint, heldCenters func() map[string]Vec3, ev RawInputMsg) {
	centers := heldCenters()
	camCenters := make(map[string]Camera.Vec3, len(centers))
	polarCenters := make(map[string]polar.Vec3, len(centers))
	for id, c := range centers {
		camCenters[id] = Camera.Vec3(c)
		polarCenters[id] = polar.Vec3(c)
	}

	pivot := Camera.FocusAhead(vp, camCenters)
	g.RotPivot = Vec3(pivot)

	eye := Camera.EyeOf(vp)
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	ndcX, ndcY, _ := Camera.ProjectNDC(pivot, eye, basis, g.Fov, g.Rect.Aspect())
	g.RotCx = ((ndcX+1)/2)*g.Rect.Width + g.Rect.Left
	g.RotCy = ((-ndcY+1)/2)*g.Rect.Height + g.Rect.Top

	_, csRadius := polar.ContentSphereOf(polarCenters)
	pivotDist := eye.Sub(pivot).Length()
	fovRad := g.Fov * math.Pi / 180
	rpx := (g.Rect.Height / 2) / math.Tan(fovRad/2)
	if pivotDist > 0 {
		rpx *= csRadius / pivotDist
	}
	g.RotPxPerRad = math.Max(rpx*(2/math.Pi), 1)
}
