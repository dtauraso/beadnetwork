package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

func RingPlaneDirection(axisPhi, axisTheta, t float64) polar.Polar {
	axis := polar.Polar2cart(polar.Polar{R: 1, Phi: axisPhi, Theta: axisTheta})

	ref := vec3{X: 0, Y: 1, Z: 0}
	u := ref.Sub(axis.Scale(ref.Dot(axis)))
	if u.Length() < 1e-9 {
		ref = vec3{X: 1, Y: 0, Z: 0}
		u = ref.Sub(axis.Scale(ref.Dot(axis)))
	}
	u = u.Normalize()
	v := axis.Cross(u)

	dir := u.Scale(math.Cos(t)).Add(v.Scale(math.Sin(t)))
	return polar.Cart2polar(dir)
}
