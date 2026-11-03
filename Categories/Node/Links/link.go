package Links

import "math"

type Start struct{ Phi, Theta, R float64 }

type Ticks struct{ Phi, Theta, R int }

type Lock struct{ Phi, Theta, R int }

type Link struct {
	From, To  string
	Start     Start
	Delta     Ticks
	Increment Ticks
	Lock      Lock
	Locking   bool
	Picked    bool
	Order     int
}

const NewLinkOrder = 1 << 20

const (
	KPhi   = "phi"
	KTheta = "theta"
	KR     = "r"
)

func NewLink(from, to string) *Link {
	return &Link{From: from, To: to, Start: Start{R: 2}, Increment: Ticks{Phi: 1, Theta: 1, R: 1}}
}

func (l *Link) Name() string { return l.From + "→" + l.To }

func (l *Link) Dir(s int) Start {
	add := func(a float64, d int) float64 { return Sum(s, a+float64(d)) }
	return Start{Phi: add(l.Start.Phi, l.Delta.Phi), Theta: add(l.Start.Theta, l.Delta.Theta), R: l.Start.R + float64(l.Delta.R)}
}

func (a *Ticks) at(k string) *int {
	if k == KR {
		return &a.R
	}
	if k == KTheta {
		return &a.Theta
	}
	return &a.Phi
}

func (k *Lock) at(name string) *int {
	if name == KR {
		return &k.R
	}
	if name == KTheta {
		return &k.Theta
	}
	return &k.Phi
}

func (l *Link) SetStart(s int, k string, v float64) {
	switch k {
	case KR:
		l.Start.R = v
	case KTheta:
		l.Start.Theta = Norm(s, v)
	default:
		l.Start.Phi = Norm(s, v)
	}
}

func (l *Link) Step(s int, k string, sign int) {
	d := l.Delta.at(k)
	*d += sign * *l.Increment.at(k)
}

func (l *Link) ScaleAngles(k float64) {
	l.Start.Phi *= k
	l.Start.Theta *= k
	l.Delta.Phi = int(math.Round(float64(l.Delta.Phi) * k))
	l.Delta.Theta = int(math.Round(float64(l.Delta.Theta) * k))
}

func Measure(f Frame, s int, was Start, p, t Vec3) Start {
	was = Start{Phi: Turn(s, was.Phi), Theta: Turn(s, was.Theta), R: was.R}
	v := t.Sub(p)
	n := v.Length()
	tick := 2 * math.Pi / float64(TurnTicks*s)
	if n < 1e-9 {
		return Start{Phi: Signed(s, was.Phi), Theta: Signed(s, was.Theta)}
	}
	side := math.Hypot(v.X, v.Z)
	phi := math.Atan2(side, v.Y) / tick
	theta := was.Theta
	if side >= 1e-9*n {
		theta = math.Atan2(v.Z, v.X) / tick
	}
	gap := func(a, b float64) float64 {
		d := Wrap(s, a-b)
		return math.Min(d, float64(TurnTicks*s)-d)
	}
	off := func(c Start) float64 { return gap(c.Phi, was.Phi) + gap(c.Theta, was.Theta) }
	half := float64(TurnTicks * s / 2)
	best := Start{Phi: phi, Theta: theta}
	for _, c := range []Start{{Phi: -phi, Theta: theta + half}, {Phi: theta, Theta: theta}, {Phi: theta + half, Theta: theta + half}} {
		d := f.Vec(s, Start{Phi: Signed(s, c.Phi), Theta: Signed(s, c.Theta), R: n / f.Rho}).Sub(v)
		if d.Length() < 1e-6*n && off(c) < off(best) {
			best = c
		}
	}
	return Start{Phi: Signed(s, best.Phi), Theta: Signed(s, best.Theta), R: n / f.Rho}
}

func (l *Link) Remeasure(f Frame, s int, p, t Vec3) {
	v := Measure(f, s, l.Dir(s), p, t)
	sub := func(a float64, d int) float64 {
		if f := Sum(s, a-float64(d)); math.Abs(Sum(s, f+float64(d))-a) < 1e-9 {
			return f
		}
		return a - float64(d)
	}
	l.Start = Start{Phi: sub(v.Phi, l.Delta.Phi), Theta: sub(v.Theta, l.Delta.Theta), R: v.R - float64(l.Delta.R)}
}
