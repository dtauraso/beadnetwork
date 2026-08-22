package polar

import (
	"math"
)

type Polar struct {
	R     float64
	Phi   float64
	Theta float64
}

func Compose(p, q Polar) Polar {
	return Polar{R: p.R + q.R, Phi: p.Phi + q.Phi, Theta: wrapTurn(p.Theta + q.Theta)}
}

func Between(from, to Polar) Polar {
	return Polar{R: to.R - from.R, Phi: to.Phi - from.Phi, Theta: wrapTurn(to.Theta - from.Theta)}
}

func SnapDeltaTheta(d Polar) Polar {
	return Polar{R: d.R, Phi: d.Phi, Theta: wrapTurn(math.Round(d.Theta/math.Pi) * math.Pi)}
}

func (p Polar) Neg() Polar {
	return Polar{R: -p.R, Phi: -p.Phi, Theta: wrapTurn(-p.Theta)}
}

func wrapTurn(a float64) float64 {
	const twoPi = 2 * math.Pi
	for a > math.Pi {
		a -= twoPi
	}
	for a <= -math.Pi {
		a += twoPi
	}
	return a
}

func Polar2cart(p Polar) Vec3 {
	st := math.Sin(p.Phi)
	return Vec3{
		X: p.R * st * math.Cos(p.Theta),
		Y: p.R * math.Cos(p.Phi),
		Z: p.R * st * math.Sin(p.Theta),
	}
}

func WorldAxisPole() (phi, theta float64) {
	return 0, 0
}

func phiOf(v Vec3) float64 {
	return math.Atan2(math.Hypot(v.X, v.Z), v.Y)
}

func Cart2polar(v Vec3) Polar {
	return Polar{R: v.Length(), Phi: phiOf(v), Theta: math.Atan2(v.Z, v.X)}
}

func Cart2polarAtTheta(v Vec3, theta float64) Polar {
	axial := v.X*math.Cos(theta) + v.Z*math.Sin(theta)
	return Polar{R: math.Hypot(axial, v.Y), Phi: math.Atan2(axial, v.Y), Theta: theta}
}
