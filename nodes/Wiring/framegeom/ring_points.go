package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
)

const RingCenterlineSteps = 32

func RingCenterlinePoints(center vec3, r, ringAxisPhi, ringAxisTheta float64) []vec3 {
	axis := camera.AnglesToWorldOffset(1, ringAxisPhi, ringAxisTheta)
	step := 2 * math.Pi / float64(RingCenterlineSteps)
	pts := make([]vec3, RingCenterlineSteps)
	for k := 0; k < RingCenterlineSteps; k++ {
		phi := float64(k) * step
		base := vec3{X: r * math.Sin(phi), Y: r * math.Cos(phi), Z: 0}
		pts[k] = center.Add(rotateFromWorldZ(base, axis))
	}
	return pts
}

func rotateFromWorldZ(v, target vec3) vec3 {
	z := vec3{X: 0, Y: 0, Z: 1}
	t := target.Normalize()
	cosA := z.Dot(t)
	if cosA > 1-1e-9 {
		return v
	}
	if cosA < -1+1e-9 {
		return vec3{X: v.X, Y: -v.Y, Z: -v.Z}
	}
	k := z.Cross(t).Normalize()
	sinA := math.Sqrt(1 - cosA*cosA)
	return v.Scale(cosA).Add(k.Cross(v).Scale(sinA)).Add(k.Scale(k.Dot(v) * (1 - cosA)))
}
