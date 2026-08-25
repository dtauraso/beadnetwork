package Camera

import (
	"math"
)

func PlaneSlide(b CamBasis, r, angle, worldPerPixel float64) Vec3 {
	return b.RefX.Scale(r * math.Cos(angle) * worldPerPixel).
		Add(b.RefY.Scale(r * math.Sin(angle) * worldPerPixel))
}

func DeltaToPolar(dx, dy float64) (r, angle float64) {
	return math.Hypot(dx, dy), math.Atan2(dy, dx)
}

func PanDisplacementPolar(pos, up Dir, dx, dy, worldPerPixel float64) Vec3 {
	r, angle := DeltaToPolar(dx, -dy)
	return PlaneSlide(BasisFromViewpoint(pos, up), r, angle, worldPerPixel)
}
