package Camera

import (
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/spatial"
)

func AnglesToWorldOffset(r, phi, theta float64) spatial.Vec3 {
	return polar.Polar2cart(polar.Polar{R: r, Phi: phi, Theta: theta})
}

func WorldDirToAngles(v spatial.Vec3) Dir {
	p := polar.Cart2polar(v)
	return Dir{Phi: p.Phi, Theta: p.Theta}
}
