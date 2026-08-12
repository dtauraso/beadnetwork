package geom

import "math"

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

func ToWorldDir(b CamBasis, q PolarDir) vec3 {
	s := math.Sin(q.Phi)
	equator := b.RefX.Scale(math.Cos(q.Theta)).Add(b.RefY.Scale(math.Sin(q.Theta)))
	return b.Pole.Scale(math.Cos(q.Phi)).Add(equator.Scale(s))
}

func PlaneSlide(b CamBasis, r, angle, worldPerPixel float64) vec3 {
	return b.RefX.Scale(r * math.Cos(angle) * worldPerPixel).
		Add(b.RefY.Scale(r * math.Sin(angle) * worldPerPixel))
}

func DeltaToPolar(dx, dy float64) (r, angle float64) {
	return math.Hypot(dx, dy), math.Atan2(dy, dx)
}

func PanDisplacementPolar(pos, up Dir, dx, dy, worldPerPixel float64) vec3 {
	r, bearing := DeltaToPolar(dx, -dy)
	_, psiUp := AzimuthFrom(pos, up)
	d := FromAxisFrame(pos, math.Pi/2, psiUp-math.Pi/2+bearing)
	return Polar2cart(Polar{R: r * worldPerPixel, Theta: d.Theta, Phi: d.Phi})
}
