package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Polar/polar"
)

func CanonicalTorusSurfacePoints(a float64, nu, nv int) []Vec3 {
	const rho = 1.0

	pts := make([]Vec3, 0, nu*nv)
	for j := 0; j < nv; j++ {
		v := float64(j) * (2 * math.Pi / float64(nv))

		w := rho + a*math.Cos(v)
		R := math.Sqrt(rho*rho + a*a + 2*rho*a*math.Cos(v))
		psi := math.Atan2(a*math.Sin(v), w)
		cosPsi := math.Cos(psi)
		sinPsi := math.Sin(psi)

		for i := 0; i < nu; i++ {
			u := float64(i) * (2 * math.Pi / float64(nu))
			sinU := math.Sin(u)
			cosU := math.Cos(u)

			phi := math.Atan2(math.Sqrt(cosPsi*cosPsi*sinU*sinU+sinPsi*sinPsi), cosPsi*cosU)
			theta := math.Atan2(sinPsi, cosPsi*sinU)

			pts = append(pts, Vec3(polar.Polar2cart(polar.Polar{R: R, Phi: phi, Theta: theta})))
		}
	}
	return pts
}

func FlattenPoints(pts []Vec3) []float32 {
	flat := make([]float32, 0, len(pts)*3)
	for _, p := range pts {
		flat = append(flat, float32(p.X), float32(p.Y), float32(p.Z))
	}
	return flat
}

func ringAxisBasis(axis Vec3) (bx, by, bz Vec3) {
	z := Vec3{X: 0, Y: 0, Z: 1}
	t := axis.Normalize()
	cosA := z.Dot(t)
	if cosA > 1-1e-9 {
		return Vec3{X: 1, Y: 0, Z: 0}, Vec3{X: 0, Y: 1, Z: 0}, Vec3{X: 0, Y: 0, Z: 1}
	}
	if cosA < -1+1e-9 {
		return Vec3{X: 1, Y: 0, Z: 0}, Vec3{X: 0, Y: -1, Z: 0}, Vec3{X: 0, Y: 0, Z: -1}
	}
	k := z.Cross(t).Normalize()
	sinA := math.Sqrt(1 - cosA*cosA)
	rot := func(v Vec3) Vec3 {
		return v.Scale(cosA).Add(k.Cross(v).Scale(sinA)).Add(k.Scale(k.Dot(v) * (1 - cosA)))
	}
	return rot(Vec3{X: 1, Y: 0, Z: 0}), rot(Vec3{X: 0, Y: 1, Z: 0}), rot(Vec3{X: 0, Y: 0, Z: 1})
}

func RingInstanceMatrixColumnMajor(center Vec3, radius, axisPhi, axisTheta float64) [16]float32 {
	axis := Camera.AnglesToWorldOffset(1, axisPhi, axisTheta)
	bx, by, bz := ringAxisBasis(Vec3(axis))
	bx = bx.Scale(radius)
	by = by.Scale(radius)
	bz = bz.Scale(radius)
	return [16]float32{
		float32(bx.X), float32(bx.Y), float32(bx.Z), 0,
		float32(by.X), float32(by.Y), float32(by.Z), 0,
		float32(bz.X), float32(bz.Y), float32(bz.Z), 0,
		float32(center.X), float32(center.Y), float32(center.Z), 1,
	}
}
