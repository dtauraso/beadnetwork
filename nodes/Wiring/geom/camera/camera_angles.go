package camera

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

// Both of these are the camera's SPELLING of the one conversion — loose angles
// in and out instead of a Polar — not a second copy of the formula. The trig
// lives once, in the polar package.

func AnglesToWorldOffset(r, theta, phi float64) vec3 {
	return polar.Polar2cart(polar.Polar{R: r, Theta: theta, Phi: phi})
}

func WorldDirToAngles(v vec3) Dir {
	p := polar.Cart2polar(v)
	return Dir{Theta: p.Theta, Phi: p.Phi}
}
