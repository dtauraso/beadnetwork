package Camera

import (
	"math"

	"github.com/dtauraso/wirefold/src/spatial"
)

type PolarDir struct {
	Theta float64
	Phi   float64
}

func ScreenToPolar(dxFromCenter, dyFromCenter, scale float64) PolarDir {
	return PolarDir{
		Phi:   math.Hypot(dxFromCenter, dyFromCenter) / scale,
		Theta: math.Atan2(-dyFromCenter, dxFromCenter),
	}
}

func ToWorldDir(b CamBasis, q PolarDir) spatial.Vec3 {
	s := math.Sin(q.Phi)
	equator := b.RefX.Scale(math.Cos(q.Theta)).Add(b.RefY.Scale(math.Sin(q.Theta)))
	return b.Pole.Scale(math.Cos(q.Phi)).Add(equator.Scale(s))
}

func PlaneSlide(b CamBasis, r, angle, worldPerPixel float64) spatial.Vec3 {
	return b.RefX.Scale(r * math.Cos(angle) * worldPerPixel).
		Add(b.RefY.Scale(r * math.Sin(angle) * worldPerPixel))
}

func DeltaToPolar(dx, dy float64) (r, angle float64) {
	return math.Hypot(dx, dy), math.Atan2(dy, dx)
}

func PanDisplacementPolar(pos, up Dir, dx, dy, worldPerPixel float64) spatial.Vec3 {
	r, angle := DeltaToPolar(dx, -dy)
	return PlaneSlide(BasisFromViewpoint(pos, up), r, angle, worldPerPixel)
}
