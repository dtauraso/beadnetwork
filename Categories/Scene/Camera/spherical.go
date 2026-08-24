package Camera

import (
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Vector/polar"
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

	if cross.Length() < 1e-12 && dot < 0 {
		panic(fmt.Sprintf(
			"camera.ArcBetween: asked for the rotation from (phi=%.6f,theta=%.6f) to its exact "+
				"opposite (phi=%.6f,theta=%.6f) — every perpendicular axis performs that 180-degree "+
				"turn and none is more correct, so there is no rotation to name. These two are "+
				"consecutive samples of ONE pointer during a drag (gesture handlers -> "+
				"Viewpoint.Orbit), which cannot reach the antipode between two events; whatever "+
				"produced this pair is not sampling a drag.",
			from.Phi, from.Theta, to.Phi, to.Theta))
	}

	return Rot{Axis: vecToDir(cross), Angle: math.Atan2(cross.Length(), dot)}
}
