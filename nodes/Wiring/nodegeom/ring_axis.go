package nodegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

func PoleContainingEdge(poleTheta, polePhi float64, selfCenter, partnerCenter vec3) (theta, phi float64, ok bool) {
	delta := partnerCenter.Sub(selfCenter)
	if delta.Length() < 1e-9 {
		return 0, 0, false
	}
	dir := delta.Normalize()
	pole := camera.AnglesToWorldOffset(1, poleTheta, polePhi)
	projected := pole.Sub(dir.Scale(pole.Dot(dir)))
	if projected.Length() < 1e-6 {
		return 0, 0, false
	}
	u := projected.Normalize()
	return math.Acos(polar.Clamp(u.Y, -1, 1)), math.Atan2(u.Z, u.X), true
}

func TorusDefaultAxisAngles() (theta, phi float64) {
	return math.Pi / 2, math.Pi / 2
}

func UprightRingAxis(selfCenter, partnerCenter vec3) (theta, phi float64, ok bool) {
	delta := partnerCenter.Sub(selfCenter)
	if delta.Length() < 1e-9 {
		return 0, 0, false
	}
	n := delta.Normalize().Cross(vec3{X: 0, Y: 1, Z: 0})
	if n.Length() < 1e-6 {
		return 0, 0, false
	}
	u := n.Normalize()
	return math.Acos(polar.Clamp(u.Y, -1, 1)), math.Atan2(u.Z, u.X), true
}
