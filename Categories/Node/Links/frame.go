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

func (f Frame) GridNear(p Vec3, was Place) Place {
	j := f.GridAt(p)
	if n := p.Length(); n >= 1e-9 && math.Hypot(p.X, p.Z) < 1e-9*n {
		j.Theta = was.Theta
		return j
	}
	tt := float64(f.Ttheta)
	gap := func(a float64) float64 {
		d := math.Mod(math.Mod(a-was.Theta, tt)+tt, tt)
		return math.Min(d, tt-d)
	}
	o := Place{R: j.R, Phi: float64(f.Tphi) - j.Phi, Theta: math.Mod(j.Theta+tt/2, tt)}
	if gap(o.Theta) < gap(j.Theta)-1e-9 {
		return o
	}
	return j
}

func (f Frame) Vec(s int, t Angles) Vec3 {
	return f.cart(s, t.R, Wrap(s, t.Phi), Wrap(s, t.Theta))
}

func (f Frame) Sum(l Link, ref Turn) Angles {
	d, a := l.Dir(), ref.In(1)
	return Measure(f, Angles{Phi: d.Phi + a.Phi, Theta: d.Theta + a.Theta}, Vec3{}, f.LinkVec(l, ref))
}

func Normalized(l Link, s int) Link {
	l.Start.Signed()
	l.SetS(s)
	return l
}

func (f Frame) cart(s int, r, phi, theta float64) Vec3 {
	ang := func(x float64) float64 { return x * 2 * math.Pi / float64(TurnTicks*s) }
	v := Cart(r*f.Rho, ang(phi), ang(theta))
	if phi > half(s) && math.Abs(phi-theta) < 1e-9 {
		v.Z = -v.Z
	}
	return v
}

type Ways struct{ Phi, Theta int }

const (
	WayNone = 0
	WayCW   = 1
	WayCCW  = -1
)

func signOf(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

func (l *Link) MarkWays() {
	l.Way = Ways{Phi: signOf(l.Delta.Phi), Theta: signOf(l.Delta.Theta)}
}

func (s *Start) otherPair() {
	per := s.per()
	s.Tick = Ticks{Phi: WrapTicks(per, -s.Tick.Phi), Theta: WrapTicks(per, s.Tick.Theta+TurnTicks*per/2)}
}

func (g Grid) thetaPressed(from string) bool {
	for _, c := range g.Cols {
		if in, ok := c.InFrom(from); ok && in.Link.Way.Theta != WayNone {
			return true
		}
	}
	return false
}

func (p *Peer) keepTheta(from string, l *Link) bool {
	if p.grid.thetaPressed(from) {
		return false
	}
	d := float64(l.Start.Tick.Theta)/float64(TurnTicks*l.Start.per()) - l.knownTheta
	if d -= math.Round(d); math.Abs(d) < 0.25 {
		return false
	}
	l.Start.otherPair()
	return true
}
