package Camera

import (
	"github.com/dtauraso/wirefold/src/Polar/polar"
)

func AnglesToWorldOffset(r, phi, theta float64) vec3 {
	return polar.Polar2cart(polar.Polar{R: r, Phi: phi, Theta: theta})
}

func WorldDirToAngles(v vec3) Dir {
	p := polar.Cart2polar(v)
	return Dir{Phi: p.Phi, Theta: p.Theta}
}
