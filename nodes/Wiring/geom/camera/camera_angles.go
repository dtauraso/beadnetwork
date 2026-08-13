package camera

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

// Both of these are the camera's SPELLING of the one conversion — loose angles
// in and out instead of a Polar — not a second copy of the formula. The trig
// lives once, in the polar package.

func AnglesToWorldOffset(r, phi, theta float64) vec3 {
	return polar.Polar2cart(polar.Polar{R: r, Phi: phi, Theta: theta})
}

func WorldDirToAngles(v vec3) Dir {
	p := polar.Cart2polar(v)
	return Dir{Phi: p.Phi, Theta: p.Theta}
}
