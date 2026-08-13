package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

func PoleContainingEdge(polePhi, poleTheta float64, selfCenter, partnerCenter vec3) (phi, theta float64, ok bool) {
	delta := partnerCenter.Sub(selfCenter)
	if delta.Length() < 1e-9 {
		return 0, 0, false
	}
	dir := delta.Normalize()
	pole := camera.AnglesToWorldOffset(1, polePhi, poleTheta)
	projected := pole.Sub(dir.Scale(pole.Dot(dir)))
	if projected.Length() < 1e-6 {
		return 0, 0, false
	}
	u := projected.Normalize()
	p := polar.Cart2polar(u)
	return p.Phi, p.Theta, true
}

func TorusDefaultAxisAngles() (phi, theta float64) {
	return math.Pi / 2, math.Pi / 2
}

func UprightRingAxis(selfCenter, partnerCenter vec3) (phi, theta float64, ok bool) {
	delta := partnerCenter.Sub(selfCenter)
	if delta.Length() < 1e-9 {
		return 0, 0, false
	}
	n := delta.Normalize().Cross(vec3{X: 0, Y: 1, Z: 0})
	if n.Length() < 1e-6 {
		return 0, 0, false
	}
	u := n.Normalize()
	p := polar.Cart2polar(u)
	return p.Phi, p.Theta, true
}
