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
	ang := func(x float64) float64 { return x * 2 * math.Pi / float64(TurnTicks*s) }
	v := Cart(t.R*f.Rho, ang(t.Phi), ang(t.Theta))
	if p := Wrap(s, t.Phi); p > float64(TurnTicks*s/2) && math.Abs(p-Wrap(s, t.Theta)) < 1e-9 {
		v.Z = -v.Z
	}
	return v
}
