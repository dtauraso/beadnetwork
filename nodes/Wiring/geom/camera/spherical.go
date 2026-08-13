package camera

import (
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

type Dir struct {
	Theta float64
	Phi   float64
}

type Rot struct {
	Axis  Dir
	Angle float64
}

// Every function here converts to a vector at the boundary, does its work as
// vectors, and converts back at the boundary. No angle is ever produced by
// adding or subtracting angles.
//
// That is what removed WrapPi. An angle built by arithmetic can land outside
// the range atan2 hands back, so it had to be folded in; an angle that only
// ever comes OUT of Cart2polar is already in that range, and there is nothing
// left to fold. The spherical-trig identities this file used to carry — law of
// cosines, bearing formula — are gone with it: they were the angle arithmetic.

func dirToVec(d Dir) vec3 {
	return polar.Polar2cart(polar.Polar{R: 1, Theta: d.Theta, Phi: d.Phi})
}

func vecToDir(v vec3) Dir {
	p := polar.Cart2polar(v)
	return Dir{Theta: p.Theta, Phi: p.Phi}
}

// RotateDir turns p about axis by angle, as a rotation of the vector itself
// (Rodrigues): v·cos + (k×v)·sin + k·(k·v)·(1−cos).
//
// It used to measure p's bearing about the axis, add the angle to it, and
// rebuild the direction from that bearing. Rotating the vector needs no bearing
// and therefore no reference direction — which is what removed the pole case,
// where a bearing has no meaning because every direction from the pole is south.
func RotateDir(p, axis Dir, angle float64) Dir {
	v, k := dirToVec(p), dirToVec(axis)
	cos, sin := math.Cos(angle), math.Sin(angle)

	return vecToDir(v.Scale(cos).
		Add(k.Cross(v).Scale(sin)).
		Add(k.Scale(k.Dot(v) * (1 - cos))))
}

// ArcBetween is the rotation carrying `from` to `to`: the axis is perpendicular
// to both, and the angle is the distance between them.
//
// The parallel case needs no handling. The cross product is zero, so the angle
// is zero, and a zero-degree turn is the identity about any axis — including
// the one the zero vector converts to. That case is COMMON: a pointer-move
// event with no movement produces it.
//
// The antiparallel case is not handled because it cannot be produced, and the
// code no longer pretends otherwise. It asserts instead.
func ArcBetween(from, to Dir) Rot {
	fv, tv := dirToVec(from), dirToVec(to)
	cross := fv.Cross(tv)
	dot := fv.Dot(tv)

	if cross.Length() < 1e-12 && dot < 0 {
		panic(fmt.Sprintf(
			"camera.ArcBetween: asked for the rotation from (theta=%.6f,phi=%.6f) to its exact "+
				"opposite (theta=%.6f,phi=%.6f) — every perpendicular axis performs that 180-degree "+
				"turn and none is more correct, so there is no rotation to name. These two are "+
				"consecutive samples of ONE pointer during a drag (gesture handlers -> "+
				"Viewpoint.Orbit), which cannot reach the antipode between two events; whatever "+
				"produced this pair is not sampling a drag.",
			from.Theta, from.Phi, to.Theta, to.Phi))
	}

	return Rot{Axis: vecToDir(cross), Angle: math.Atan2(cross.Length(), dot)}
}
