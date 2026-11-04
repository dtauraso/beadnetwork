package Links

import "math"

const pathSteps = 8

type Path struct {
	Pts  []Vec3
	Ends []int
}

func (f Frame) Arcs(s int, l Link) (phi, theta Path) {
	d := l.Dir(s)
	pp := float64(l.Delta.Phi)
	phi = f.path(s, l.Start.Phi, 0, func(v float64) Vec3 { return f.cart(s, 1, v, 0) })
	phi = phi.join(f.path(s, 0, l.Delta.Phi, func(v float64) Vec3 { return PoleAxesAt(s, v, 0).Out(f.cart(s, 1, d.Phi, 0)) }), l.Delta.Phi)
	theta = f.path(s, l.Start.Theta, 0, func(v float64) Vec3 { return PoleAxesAt(s, pp, 0).Out(f.cart(s, 1, d.Phi, v)) })
	theta = theta.join(f.path(s, 0, l.Delta.Theta, func(v float64) Vec3 { return PoleAxesAt(s, pp, v).Out(f.cart(s, 1, d.Phi, d.Theta)) }), l.Delta.Theta)
	return phi, theta
}

func (a Path) join(b Path, ticks int) Path {
	if ticks == 0 {
		return a
	}
	out := Path{Pts: append(append([]Vec3{}, a.Pts...), b.Pts[1:]...), Ends: append([]int{}, a.Ends...)}
	for _, e := range b.Ends {
		out.Ends = append(out.Ends, len(a.Pts)+e-1)
	}
	return out
}

func presses(s int, start float64, delta int) []float64 {
	vs := []float64{0, Wrap(s, start)}
	sign := 1
	if delta < 0 {
		sign = -1
	}
	v := vs[1]
	for n := delta; n != 0; n -= sign {
		v = Wrap(s, v+float64(sign))
		vs = append(vs, v)
	}
	return vs
}

func (f Frame) path(s int, start float64, delta int, at func(turn float64) Vec3) Path {
	vs := presses(s, start, delta)
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
