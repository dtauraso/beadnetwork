package Camera

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

type Dir struct {
	Phi   float64
	Theta float64
}

type Rot struct {
	Axis  Dir
	Angle float64
}

func dirToVec(d Dir) Vec3 {
	return Vec3(polar.Polar2cart(polar.Polar{R: 1, Phi: d.Phi, Theta: d.Theta}))
}

func vecToDir(v Vec3) Dir {
	p := polar.Cart2polar(polar.Vec3(v))
	return Dir{Phi: p.Phi, Theta: p.Theta}
}

func rotateVecAbout(v, k Vec3, angle float64) Vec3 {
	cos, sin := math.Cos(angle), math.Sin(angle)

	return v.Scale(cos).
		Add(k.Cross(v).Scale(sin)).
		Add(k.Scale(k.Dot(v) * (1 - cos)))
}

func AngleAboutAxis(from, to, axis Dir) float64 {
	av := dirToVec(axis)
	fv, tv := dirToVec(from), dirToVec(to)

	fp := fv.Sub(av.Scale(av.Dot(fv)))
	tp := tv.Sub(av.Scale(av.Dot(tv)))

	return math.Atan2(av.Dot(fp.Cross(tp)), fp.Dot(tp))
}

func RotateDir(p, axis Dir, angle float64) Dir {
	v, k := dirToVec(p), dirToVec(axis)
	cos, sin := math.Cos(angle), math.Sin(angle)

	return vecToDir(v.Scale(cos).
		Add(k.Cross(v).Scale(sin)).
		Add(k.Scale(k.Dot(v) * (1 - cos))))
}

func ArcBetween(from, to Dir) Rot {
	fv, tv := dirToVec(from), dirToVec(to)
	cross := fv.Cross(tv)
	dot := fv.Dot(tv)

	if cross.Length() < 1e-12 {
		if dot >= 0 {
			return Rot{Axis: from, Angle: 0}
		}
		return Rot{Axis: vecToDir(anyPerpendicularTo(fv)), Angle: math.Pi}
	}

	return Rot{Axis: vecToDir(cross), Angle: math.Atan2(cross.Length(), dot)}
}

func anyPerpendicularTo(v Vec3) Vec3 {
	away := Vec3{X: 1}
	if math.Abs(v.X) > 0.9 {
		away = Vec3{Y: 1}
	}
	return v.Cross(away).Normalize()
}
