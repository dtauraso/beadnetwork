package Links

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Vectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Vec3 = Vectors.Vec3

type Place = polarindex.Place

const TurnTicks = 12

type Frame struct {
	Rho    float64
	Cr     float64
	Tphi   int
	Ttheta int
}

func Cart(r, phi, theta float64) Vec3 {
	return Vec3{X: r * math.Sin(phi) * math.Cos(theta), Y: r * math.Cos(phi), Z: r * math.Sin(phi) * math.Sin(theta)}
}

func Wrap(s int, t float64) float64 {
	turn := float64(TurnTicks * s)
	return math.Mod(math.Mod(t, turn)+turn, turn)
}

func WrapTicks(s, t int) int {
	turn := TurnTicks * s
	return (t%turn + turn) % turn
}

func half(s int) float64 { return float64(TurnTicks * s / 2) }

func Signed(s int, turn float64) float64 {
	six := half(s)
	w := Wrap(s, turn)
	switch {
	case w == 0:
		return 0
	case w <= six:
		return w
	default:
		return -(w - six)
	}
}

func Turn(s int, signed float64) float64 {
	six := half(s)
	switch {
	case signed == 0:
		return 0
	case signed > 0:
		return signed
	case signed == -six:
		return 2 * six
	default:
		return six - signed
	}
}

func Norm(s int, signed float64) float64 {
	if signed >= -half(s) && signed <= half(s) {
		return signed
	}
	return Signed(s, signed)
}

func Up(s int, left bool, v float64) float64 {
	six := half(s)
	switch {
	case left && v == six:
		return six - 1
	case v >= 0 && v+1 <= six:
		return v + 1
	case v >= 0:
		return -(v + 1 - six)
	case v+1 <= -1:
		return v + 1
	default:
		return six - (v + 1)
	}
}

func Down(s int, left bool, v float64) float64 {
	six := half(s)
	switch {
	case left && v == six:
		return -1
	case v >= 0 && v-1 >= 0:
		return v - 1
	case v >= 0:
		return -six + (v - 1 + 1)
	case v-1 >= -six:
		return v - 1
	default:
		return -six - (v - 1) - 1
	}
}

func unUp(s int, left bool, v float64) float64 {
	six := half(s)
	switch {
	case left && v == six-1:
		return six
	case v > 0:
		return v - 1
	case v < 0 && v-1 >= -six:
		return v - 1
	default:
		return v
	}
}

func unDown(s int, left bool, v float64) float64 {
	six := half(s)
	switch {
	case left && v == -1:
		return six
	case v >= 0 && v+1 <= six:
		return v + 1
	case v < 0 && v+1 <= -1:
		return v + 1
	default:
		return v
	}
}

func Presses(s int, v float64, n int) float64 { return PressesIn(s, v < 0, v, n) }

func PressesIn(s int, left bool, v float64, n int) float64 {
	for ; n > 0; n-- {
		v = Up(s, left, v)
	}
	for ; n < 0; n++ {
		v = Down(s, left, v)
	}
	return v
}

func Unpress(s int, left bool, v float64, n int) float64 {
	for ; n > 0; n-- {
		v = unUp(s, left, v)
	}
	for ; n < 0; n++ {
		v = unDown(s, left, v)
	}
	return v
}

func whole(x float64) bool { return math.Abs(x-math.Round(x)) < 1e-6 }

func (f Frame) cellPhi() float64   { return 2 * math.Pi / float64(f.Tphi) }
func (f Frame) cellTheta() float64 { return 2 * math.Pi / float64(f.Ttheta) }

func (f Frame) At(i Place) Vec3 {
	return Cart(i.R*f.Cr, i.Phi*f.cellPhi(), i.Theta*f.cellTheta())
}

func (f Frame) IndexOf(p Vec3) Place {
	r := p.Length()
	if r < 1e-9 {
		return Place{}
	}
	return Place{
		R:     r / f.Cr,
		Phi:   math.Acos(math.Max(-1, math.Min(1, p.Y/r))) / f.cellPhi(),
		Theta: math.Atan2(p.Z, p.X) / f.cellTheta(),
	}
}

func (f Frame) GridAt(p Vec3) Place {
	j := f.IndexOf(p)
	if j.Theta < -1e-9 {
		j.Theta += float64(f.Ttheta)
	} else {
		j.Theta = math.Max(0, j.Theta)
	}
	return j
}

func (f Frame) Vec(s int, t Start) Vec3 {
	return f.cart(s, t.R, Turn(s, t.Phi), Turn(s, t.Theta))
}

func (f Frame) cart(s int, r, phi, theta float64) Vec3 {
	ang := func(x float64) float64 { return x * 2 * math.Pi / float64(TurnTicks*s) }
	v := Cart(r*f.Rho, ang(phi), ang(theta))
	if phi > half(s) && math.Abs(phi-theta) < 1e-9 {
		v.Z = -v.Z
	}
	return v
}
