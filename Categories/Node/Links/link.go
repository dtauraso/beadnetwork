package Links

import "math"

type Start struct {
	Tick Ticks
	R    float64
	S    int
	One  Ones
}

type Ticks struct{ Phi, Theta int }

type Angles struct{ Phi, Theta, R float64 }

func (s Start) per() int { return max(1, s.S) }

func (s Start) Angles() Angles {
	f := float64(s.per())
	return Angles{Phi: float64(s.Tick.Phi) / f, Theta: float64(s.Tick.Theta) / f, R: s.R}
}

func (s *Start) Cut(a Angles) {
	f := float64(s.per())
	cut := func(v float64) int { return WrapTicks(s.per(), int(math.Trunc(Wrap(1, v)*f))) }
	s.Tick, s.R = Ticks{cut(a.Phi), cut(a.Theta)}, a.R
}

func (s *Start) Rescale(to int) {
	from, to := s.per(), max(1, to)
	s.Tick = Ticks{s.Tick.Phi * to / from, s.Tick.Theta * to / from}
	s.S = to
}

func (s *Start) tickOf(k string) *int {
	if k == KTheta {
		return &s.Tick.Theta
	}
	return &s.Tick.Phi
}

func (s *Start) Turn(k string, presses int) {
	t := s.tickOf(k)
	*t = WrapTicks(s.per(), *t+presses*s.per())
}

func (s *Start) SetTick(k string, n int) { *s.tickOf(k) = WrapTicks(s.per(), n) }

func (s *Start) Signed() {
	s.Tick = Ticks{signedTicks(s.per(), s.Tick.Phi), signedTicks(s.per(), s.Tick.Theta)}
}

func signedTicks(s, v int) int {
	if v < 0 {
		return WrapTicks(s, TurnTicks*s/2-v)
	}
	return WrapTicks(s, v)
}

func LegacyTicks(v float64, s int) int { return int(math.Round(v * float64(max(1, s)))) }

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
	Total     Ints
	Increment Ints
	Lock      Lock
	Way       Ways
	Picked    bool
	Order     int
	Playing   bool
	Part      float64
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

func (l *Link) Sum() Ints {
	return Ints{Phi: l.Total.Phi + l.Delta.Phi, Theta: l.Total.Theta + l.Delta.Theta, R: l.Total.R + l.Delta.R, S: l.Delta.S}
}

func (l *Link) Dir() Angles {
	at, d := l.Start.Angles(), l.Sum()
	return Angles{Phi: Wrap(1, at.Phi+float64(d.Phi)), Theta: Wrap(1, at.Theta+float64(d.Theta)), R: at.R + float64(d.R)}
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
	if carry == 0 {
		return
	}
	if k == KR {
		l.Start.R += float64(carry)
		return
	}
	l.Start.Turn(k, carry)
}

func ShownS(s int) int { return max(1, s) }

func (l *Link) SetS(s int) {
	l.Start.Rescale(s)
	l.Delta.S, l.Increment.S = s, s
}

func (l *Link) Drawn() (int, Angles) {
	p, t, d := l.Start.per(), l.Total, l.Delta
	f, part := float64(p), 1.0
	if l.Playing {
		part = l.Part
	}
	at := func(t, d int) float64 { return float64(t) + float64(d)*part }
	return p, Angles{Phi: float64(l.Start.Tick.Phi) + at(t.Phi, d.Phi)*f, Theta: float64(l.Start.Tick.Theta) + at(t.Theta, d.Theta)*f, R: l.Start.R + at(t.R, d.R)}
}

func (l *Link) Presses() int {
	n := 0
	for _, k := range []string{KPhi, KTheta, KR} {
		d, inc := *l.Delta.At(k), max(1, *l.Increment.At(k))
		n = max(n, (max(d, -d)+inc-1)/inc)
	}
	return n
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

func (l *Link) Step(k string, sign int) {
	*l.Delta.At(k) += sign * *l.Increment.At(k)
}

func Measure(f Frame, was Angles, p, t Vec3) Angles {
	const s = 1
	was = Angles{Phi: Wrap(s, was.Phi), Theta: Wrap(s, was.Theta), R: was.R}
	v := t.Sub(p)
	n := v.Length()
	tick := 2 * math.Pi / float64(TurnTicks*s)
	if n < 1e-9 {
		return Angles{Phi: was.Phi, Theta: was.Theta}
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
	off := func(c Angles) float64 { return gap(c.Phi, was.Phi) + gap(c.Theta, was.Theta) }
	half := float64(TurnTicks) * s / 2
	best := Angles{Phi: phi, Theta: theta}
	for _, c := range []Angles{{Phi: -phi, Theta: theta + half}, {Phi: theta, Theta: theta}, {Phi: theta + half, Theta: theta + half}} {
		d := f.Vec(s, Angles{Phi: c.Phi, Theta: c.Theta, R: n / f.Rho}).Sub(v)
		if d.Length() < 1e-6*n && off(c) < off(best) {
			best = c
		}
	}
	return Angles{Phi: Wrap(s, best.Phi), Theta: Wrap(s, best.Theta), R: n / f.Rho}
}

func (l *Link) Remeasure(f Frame, ref Turn, p, t Vec3) {
	l.Start.Cut(Measure(f, l.Start.Angles(), Vec3{}, ref.Unrotate(t.Sub(p))))
}
