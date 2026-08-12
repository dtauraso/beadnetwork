package gesturefsm

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func (g *GestureState) BeginSphereRotation(vp geom.Viewpoint, heldCenters func() map[string]wire.Vec3, ev inputcodec.RawInputMsg) {
	pivot := geom.FocusAhead(vp, heldCenters())
	g.RotPivot = pivot

	eye := geom.EyeOf(vp)
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	ndcX, ndcY, _ := geom.ProjectNDC(pivot, eye, basis, ev.Fov, g.Rect.Aspect())
	g.RotCx = ((ndcX+1)/2)*g.Rect.Width + g.Rect.Left
	g.RotCy = ((-ndcY+1)/2)*g.Rect.Height + g.Rect.Top

	_, csRadius := geom.ContentSphereOf(heldCenters())
	pivotDist := eye.Sub(pivot).Length()
	fovRad := ev.Fov * math.Pi / 180
	rpx := (g.Rect.Height / 2) / math.Tan(fovRad/2)
	if pivotDist > 0 {
		rpx *= csRadius / pivotDist
	}
	g.RotPxPerRad = math.Max(rpx*(2/math.Pi), 1)
}
