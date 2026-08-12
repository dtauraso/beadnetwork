package geom

import "math"

type Polar struct {
	R     float64
	Theta float64
	Phi   float64
}

func Polar2cart(p Polar) vec3 {
	st := math.Sin(p.Theta)
	return vec3{
		X: p.R * st * math.Cos(p.Phi),
		Y: p.R * math.Cos(p.Theta),
		Z: p.R * st * math.Sin(p.Phi),
	}
}

func InwardPole(p Polar) (theta, phi float64) {
	if p.R == 0 {
		return 0, 0
	}
	return math.Pi - p.Theta, WrapPi(p.Phi + math.Pi)
}

func Cart2polar(v vec3) Polar {
	r := v.Length()
	if r == 0 {
		return Polar{}
	}
	theta := math.Acos(Clamp(v.Y/r, -1, 1))
	phi := math.Atan2(v.Z, v.X)
	return Polar{R: r, Theta: theta, Phi: phi}
}

func PolarDist(a, b Polar) float64 {
	cosG := math.Cos(a.Theta)*math.Cos(b.Theta) +
		math.Sin(a.Theta)*math.Sin(b.Theta)*math.Cos(a.Phi-b.Phi)
	d2 := a.R*a.R + b.R*b.R - 2*a.R*b.R*cosG
	if d2 <= 0 {
		return 0
	}
	return math.Sqrt(d2)
}

func Clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
