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

func FromSigned(s int, v float64) float64 {
	if v < 0 {
		return Wrap(s, half(s)-v)
	}
	return Wrap(s, v)
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
	return f.cart(s, t.R, Wrap(s, t.Phi), Wrap(s, t.Theta))
}

type Axes struct{ P, U, W Vec3 }

func PoleAxes(s int, pole Ticks) Axes {
	return PoleAxesAt(s, float64(pole.Phi), float64(pole.Theta))
}

func PoleAxesAt(s int, phi, theta float64) Axes {
	ang := func(x float64) float64 { return x * 2 * math.Pi / float64(TurnTicks*s) }
	sp, cp := math.Sincos(ang(phi))
	st, ct := math.Sincos(ang(theta))
	return Axes{
		P: Vec3{X: sp * ct, Y: cp, Z: sp * st},
		U: Vec3{X: cp * ct, Y: -sp, Z: cp * st},
		W: Vec3{X: -st, Y: 0, Z: ct},
	}
}

func (a Axes) Out(v Vec3) Vec3 {
	return a.U.Scale(v.X).Add(a.P.Scale(v.Y)).Add(a.W.Scale(v.Z))
}

func (a Axes) In(v Vec3) Vec3 {
	return Vec3{X: v.Dot(a.U), Y: v.Dot(a.P), Z: v.Dot(a.W)}
}

func (f Frame) LinkVec(s int, l Link) Vec3 {
	return PoleAxes(s, l.Pole).Out(f.Vec(s, l.Dir(s)))
}

func (f Frame) cart(s int, r, phi, theta float64) Vec3 {
	ang := func(x float64) float64 { return x * 2 * math.Pi / float64(TurnTicks*s) }
	v := Cart(r*f.Rho, ang(phi), ang(theta))
	if phi > half(s) && math.Abs(phi-theta) < 1e-9 {
		v.Z = -v.Z
	}
	return v
}
