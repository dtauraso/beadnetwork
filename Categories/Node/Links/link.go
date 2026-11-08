package Links

import "math"

type Start struct {
	Phi, Theta, R float64
	S             int
}

type Ints struct{ Phi, Theta, R, S int }

type Nums struct{ Phi, Theta, R float64 }

type Lock struct{ Phi, Theta, R int }

type Link struct {
	From, To  string
	Start     Start
	Delta     Ints
	Increment Ints
	Lock      Lock
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
	return &Link{From: from, To: to, Start: Start{R: 2, S: 1}, Delta: Ints{S: 1}, Increment: Ints{Phi: 1, Theta: 1, R: 1, S: 1}}
}

func (l *Link) Name() string { return l.From + "→" + l.To }

func (l *Link) Dir() Start {
	return Start{Phi: Wrap(1, l.Start.Phi+float64(l.Delta.Phi)), Theta: Wrap(1, l.Start.Theta+float64(l.Delta.Theta)), R: l.Start.R + float64(l.Delta.R)}
}

func ShownS(s int) int { return max(1, s) }

func (l Link) Shown() Link {
	k := float64(ShownS(l.Start.S))
	l.Start.Phi, l.Start.Theta = l.Start.Phi*k, l.Start.Theta*k
	d, n := ShownS(l.Delta.S), ShownS(l.Increment.S)
	l.Delta.Phi, l.Delta.Theta = l.Delta.Phi*d, l.Delta.Theta*d
	l.Increment.Phi, l.Increment.Theta = l.Increment.Phi*n, l.Increment.Theta*n
	return l
}

func (l *Link) SetS(s int) { l.Start.S, l.Delta.S, l.Increment.S = s, s, s }

func (a *Ints) At(k string) *int {
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

func (l *Link) SetStart(k string, v float64) {
	switch k {
	case KR:
		l.Start.R = v
	case KTheta:
		l.Start.Theta = Wrap(1, v)
	default:
		l.Start.Phi = Wrap(1, v)
	}
}

func (l *Link) Step(k string, sign int) {
	*l.Delta.At(k) += sign * *l.Increment.At(k)
}

func Measure(f Frame, was Start, p, t Vec3) Start {
	const s = 1
	was = Start{Phi: Wrap(s, was.Phi), Theta: Wrap(s, was.Theta), R: was.R}
	v := t.Sub(p)
	n := v.Length()
	tick := 2 * math.Pi / float64(TurnTicks*s)
	if n < 1e-9 {
		return Start{Phi: was.Phi, Theta: was.Theta}
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
	half := float64(TurnTicks) * s / 2
	best := Start{Phi: phi, Theta: theta}
	for _, c := range []Start{{Phi: -phi, Theta: theta + half}, {Phi: theta, Theta: theta}, {Phi: theta + half, Theta: theta + half}} {
		d := f.Vec(s, Start{Phi: c.Phi, Theta: c.Theta, R: n / f.Rho}).Sub(v)
		if d.Length() < 1e-6*n && off(c) < off(best) {
			best = c
		}
	}
	return Start{Phi: Wrap(s, best.Phi), Theta: Wrap(s, best.Theta), R: n / f.Rho}
}

func (l *Link) Remeasure(f Frame, ref Turn, p, t Vec3) {
	v := Measure(f, l.Dir(), Vec3{}, ref.Unrotate(t.Sub(p)))
	l.Start = Start{Phi: Wrap(1, v.Phi-float64(l.Delta.Phi)), Theta: Wrap(1, v.Theta-float64(l.Delta.Theta)), R: v.R - float64(l.Delta.R), S: l.Start.S}
}
