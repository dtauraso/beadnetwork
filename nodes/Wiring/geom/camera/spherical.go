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

// AngularDistance is the angle between two directions on the sphere.
//
// The spherical law of cosines gives cos of that angle; its SINE is the length
// of the same two components the azimuth below is built from. Feeding both to
// atan2 removes acos's domain — the clamp existed only because rounding pushes
// the cosine past ±1 for nearly-equal directions — and is far more accurate for
// small angles, where cosine is flat and acos loses most of its digits.
func AngularDistance(a, b Dir) float64 {
	dphi := b.Phi - a.Phi
	cosD := math.Cos(a.Theta)*math.Cos(b.Theta) +
		math.Sin(a.Theta)*math.Sin(b.Theta)*math.Cos(dphi)
	sinD := math.Hypot(
		math.Sin(b.Theta)*math.Sin(dphi),
		math.Sin(a.Theta)*math.Cos(b.Theta)-math.Cos(a.Theta)*math.Sin(b.Theta)*math.Cos(dphi),
	)
	return math.Atan2(sinD, cosD)
}

func AzimuthFrom(pole, p Dir) (c, psi float64) {
	c = AngularDistance(pole, p)
	dphi := p.Phi - pole.Phi
	psi = math.Atan2(
		math.Sin(p.Theta)*math.Sin(dphi),
		math.Sin(pole.Theta)*math.Cos(p.Theta)-math.Cos(pole.Theta)*math.Sin(p.Theta)*math.Cos(dphi),
	)
	return c, psi
}

func FromAxisFrame(pole Dir, c, psi float64) Dir {
	// The same two components serve twice: atan2'd against each other they are
	// the azimuth step, and their length is sin(theta) to the cosine's cos —
	// so theta comes out of atan2 as well, with no acos and no clamp.
	tangential := math.Sin(c) * math.Sin(psi)
	meridional := math.Sin(pole.Theta)*math.Cos(c) - math.Cos(pole.Theta)*math.Sin(c)*math.Cos(psi)

	cosT := math.Cos(pole.Theta)*math.Cos(c) + math.Sin(pole.Theta)*math.Sin(c)*math.Cos(psi)
	theta := math.Atan2(math.Hypot(tangential, meridional), cosT)
	dphi := math.Atan2(tangential, meridional)

	return Dir{Theta: theta, Phi: polar.WrapPi(pole.Phi + dphi)}
}

func RotateDir(p, axis Dir, angle float64) Dir {
	c, psi := AzimuthFrom(axis, p)
	return FromAxisFrame(axis, c, psi+angle)
}

func ArcBetween(from, to Dir) Rot {
	c, psi := AzimuthFrom(from, to)
	axis := FromAxisFrame(from, math.Pi/2, psi+math.Pi/2)
	return Rot{Axis: axis, Angle: c}
}

func AngleAboutAxis(from, to, axis Dir) float64 {
	_, psiFrom := AzimuthFrom(axis, from)
	_, psiTo := AzimuthFrom(axis, to)
	return polar.WrapPi(psiTo - psiFrom)
}
