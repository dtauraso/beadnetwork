package camera

import (
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

// axisFrame is the tangent basis at p: north points along p's meridian toward
// +y, east toward increasing phi. Azimuth is measured from north toward east,
// which is the convention the callers were already written against.
//
// At the poles the meridian direction is undefined — every direction from the
// north pole is south — so phi itself supplies the reference there, which is
// what the spherical formula this replaced did implicitly.
func axisFrame(p vec3) (north, east vec3) {
	up := vec3{X: 0, Y: 1, Z: 0}
	north = up.Sub(p.Scale(up.Dot(p)))
	if north.Length() < 1e-12 {
		meridian := vec3{X: math.Cos(vecToDir(p).Phi), Y: 0, Z: math.Sin(vecToDir(p).Phi)}
		north = meridian.Sub(p.Scale(meridian.Dot(p)))
		if p.Y > 0 {
			north = north.Scale(-1)
		}
	}
	north = north.Normalize()
	east = p.Cross(up)
	if east.Length() < 1e-12 {
		east = p.Cross(north)
	}
	return north, east.Normalize()
}

// AzimuthFrom gives p's polar coordinates IN THE FRAME whose pole is `pole`:
// c is the angular distance, psi the bearing from north toward east.
func AzimuthFrom(pole, p Dir) (c, psi float64) {
	pv, tv := dirToVec(pole), dirToVec(p)
	c = math.Atan2(pv.Cross(tv).Length(), pv.Dot(tv))

	north, east := axisFrame(pv)
	perp := tv.Sub(pv.Scale(pv.Dot(tv)))
	psi = math.Atan2(perp.Dot(east), perp.Dot(north))
	return c, psi
}

// FromAxisFrame is AzimuthFrom's inverse: the direction at angular distance c
// from `pole`, on bearing psi.
func FromAxisFrame(pole Dir, c, psi float64) Dir {
	pv := dirToVec(pole)
	north, east := axisFrame(pv)

	tangent := north.Scale(math.Cos(psi)).Add(east.Scale(math.Sin(psi)))
	return vecToDir(pv.Scale(math.Cos(c)).Add(tangent.Scale(math.Sin(c))))
}

func RotateDir(p, axis Dir, angle float64) Dir {
	c, psi := AzimuthFrom(axis, p)
	return FromAxisFrame(axis, c, psi+angle)
}

// ArcBetween is the rotation carrying `from` to `to`: the axis is perpendicular
// to both, and the angle is the distance between them.
func ArcBetween(from, to Dir) Rot {
	fv, tv := dirToVec(from), dirToVec(to)
	cross := fv.Cross(tv)
	if cross.Length() < 1e-12 {
		// Parallel or antiparallel: no unique axis. Keep the old behaviour of
		// naming one perpendicular rather than failing.
		north, _ := axisFrame(fv)
		return Rot{Axis: vecToDir(fv.Cross(north)), Angle: math.Atan2(cross.Length(), fv.Dot(tv))}
	}
	return Rot{Axis: vecToDir(cross), Angle: math.Atan2(cross.Length(), fv.Dot(tv))}
}
