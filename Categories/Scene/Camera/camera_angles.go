package Camera

import (
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
)

func AnglesToWorldOffset(r, phi, theta float64) Vec3 {
	return Vec3(polar.Polar2cart(polar.Polar{R: r, Phi: phi, Theta: theta}))
}

func WorldDirToAngles(v Vec3) Dir {
	p := polar.Cart2polar(polar.Vec3(v))
	return Dir{Phi: p.Phi, Theta: p.Theta}
}
