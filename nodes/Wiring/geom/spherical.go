package geom

import (
	"math"
)

type Dir struct {
	Theta float64
	Phi   float64
}

type Rot struct {
	Axis  Dir
	Angle float64
}

func WrapPi(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a <= -math.Pi {
		a += 2 * math.Pi
	}
	return a
}

func AngularDistance(a, b Dir) float64 {
	cd := math.Cos(a.Theta)*math.Cos(b.Theta) +
		math.Sin(a.Theta)*math.Sin(b.Theta)*math.Cos(a.Phi-b.Phi)
	return math.Acos(Clamp(cd, -1, 1))
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
	cosT := Clamp(math.Cos(pole.Theta)*math.Cos(c)+math.Sin(pole.Theta)*math.Sin(c)*math.Cos(psi), -1, 1)
	theta := math.Acos(cosT)
	dphi := math.Atan2(
		math.Sin(c)*math.Sin(psi),
		math.Sin(pole.Theta)*math.Cos(c)-math.Cos(pole.Theta)*math.Sin(c)*math.Cos(psi),
	)
	return Dir{Theta: theta, Phi: WrapPi(pole.Phi + dphi)}
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
	return WrapPi(psiTo - psiFrom)
}
