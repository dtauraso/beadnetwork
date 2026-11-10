package Links

import "math"

type Start struct {
	Phi, Theta, R float64
	S             int
	One           Ones
	Tick          Ticks
}

type Ticks struct{ Phi, Theta int }

func Digits(s int) int {
	d := 0
	for p := 1; p < s; p *= 10 {
		d++
	}
	return d
}

func TenTo(s int) int {
	p := 1
	for range Digits(s) {
		p *= 10
	}
	return p
}

func (s Start) Pulled() Start {
	p := float64(TenTo(s.S))
	s.Tick = Ticks{int(math.Floor(s.Phi*p + 1e-9)), int(math.Floor(s.Theta*p + 1e-9))}
	return s
}

func (s Start) FromTick(n int) float64 {
	return float64(n) * math.Pow10(-Digits(s.S))
}

type Ones struct{ Phi, Theta, R int }

func (o *Ones) At(k string) *int {
	if k == KR {
		return &o.R
	}
	if k == KTheta {
		return &o.Theta
	}
	return &o.Phi
}

type Ints struct{ Phi, Theta, R, S int }

type Ref struct{ Phi, Theta, S int }

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
	at := l.Start
	return Start{Phi: Wrap(1, at.Phi+float64(l.Delta.Phi)), Theta: Wrap(1, at.Theta+float64(l.Delta.Theta)), R: at.R + float64(l.Delta.R)}
}

func (l *Link) OneCarry(k string, sign int) int {
	o := *l.Start.One.At(k) + sign
	if o >= ShownS(l.Start.S) {
		return 1
	}
	if o < 0 {
		return -1
	}
	return 0
}

func (l *Link) StepOne(k string, sign int) {
	carry := l.OneCarry(k, sign)
	o := l.Start.One.At(k)
	*o += sign - carry*ShownS(l.Start.S)
	if carry != 0 {
		v := map[string]float64{KPhi: l.Start.Phi, KTheta: l.Start.Theta, KR: l.Start.R}[k]
		l.SetStart(k, v+float64(carry))
	}
}

func ShownS(s int) int { return max(1, s) }

func (l *Link) SetS(s int) {
	l.Start.S, l.Delta.S, l.Increment.S = s, s, s
	l.Start = l.Start.Pulled()
}

func (l *Link) Drawn() (int, Start) {
	p := TenTo(l.Start.S)
	t := l.Start.Tick
	return p, Start{Phi: float64(t.Phi + l.Delta.Phi*p), Theta: float64(t.Theta + l.Delta.Theta*p), R: l.Start.R + float64(l.Delta.R)}
}

func Tenths(v float64) int { return int(math.Floor(v*10 + 1e-6)) }

func (s Start) Tenth() Ones { return Ones{Tenths(s.Phi), Tenths(s.Theta), Tenths(s.R)} }

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return max(a, -a)
}

func Lcm(a, b int) int { return a / gcd(a, b) * b }

func (s Start) TicksNeeded() int {
	t := s.Tenth()
	need := 1
	for _, n := range []int{t.Phi, t.Theta, t.R} {
		need = Lcm(need, 10/gcd(n, 10))
	}
	return need
}

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
	l.Start = l.Start.Pulled()
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
	l.Start = Start{Phi: Wrap(1, v.Phi-float64(l.Delta.Phi)), Theta: Wrap(1, v.Theta-float64(l.Delta.Theta)), R: v.R - float64(l.Delta.R), S: l.Start.S, One: l.Start.One}.Pulled()
}
