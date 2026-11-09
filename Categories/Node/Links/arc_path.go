package Links

import "math"

const pathSteps = 8

type Path struct {
	Pts  []Vec3
	Ends []int
}

func (f Frame) Arcs(l Link, ref Turn) (phi, theta Path) {
	const s = 1
	at := l.Start.Stepped()
	ph := Wrap(s, at.Phi+float64(l.Delta.Phi))
	phi = f.path(s, 0, at.Phi, float64(l.Delta.Phi), func(v float64) Vec3 { return ref.Rotate(f.cart(s, 1, v, 0)) })
	theta = f.path(s, 0, at.Theta, float64(l.Delta.Theta), func(v float64) Vec3 { return ref.Rotate(f.cart(s, 1, ph, v)) })
	return phi, theta
}

func presses(s int, from, start, delta float64) []float64 {
	vs := []float64{Wrap(s, from), Wrap(s, from+start)}
	v := vs[1]
	for left := delta; math.Abs(left) > 1e-9; {
		d := math.Copysign(math.Min(1, math.Abs(left)), left)
		v = Wrap(s, v+d)
		vs = append(vs, v)
		left -= d
	}
	return vs
}

func (f Frame) path(s int, from, start, delta float64, at func(turn float64) Vec3) Path {
	vs := presses(s, from, start, delta)
	turn := float64(TurnTicks * s)
	p := Path{Pts: []Vec3{at(vs[0])}}
	part := 0.0
	for i := 1; i < len(vs); i++ {
		a, b := vs[i-1], vs[i]
		end := b
		switch {
		case b-a > half(s):
			b -= turn
		case a-b > half(s):
			b += turn
		}
		for j := 1; j < pathSteps; j++ {
			p.Pts = append(p.Pts, at(Wrap(s, a+(b-a)*float64(j)/pathSteps)))
		}
		p.Pts = append(p.Pts, at(end))
		part += math.Abs(b - a)
		if (i == 1 || i == len(vs)-1) && part > 1e-9 {
			p.Ends = append(p.Ends, len(p.Pts)-1)
		}
		if i == 1 {
			part = 0
		}
	}
	return p
}
