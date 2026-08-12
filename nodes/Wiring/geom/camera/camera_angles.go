package camera

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

func AnglesToWorldOffset(r, theta, phi float64) vec3 {
	sinT := math.Sin(theta)
	return vec3{
		X: r * sinT * math.Cos(phi),
		Y: r * math.Cos(theta),
		Z: r * sinT * math.Sin(phi),
	}
}

func WorldDirToAngles(v vec3) Dir {
	d := v.Normalize()
	return Dir{
		Theta: math.Acos(polar.Clamp(d.Y, -1, 1)),
		Phi:   math.Atan2(d.Z, d.X),
	}
}
