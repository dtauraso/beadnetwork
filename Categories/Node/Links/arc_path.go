package Links

import "math"

const pathSteps = 8

type Path struct {
	Pts  []Vec3
	Ends []int
}

func (f Frame) Arcs(s int, l Link) (phi, theta Path) {
	ph := Turn(s, l.Dir(s).Phi)
	phi = f.path(s, l.Start.Phi, l.Delta.Phi, func(v float64) Vec3 { return f.cart(s, 1, v, 0) })
	theta = f.path(s, l.Start.Theta, l.Delta.Theta, func(v float64) Vec3 { return f.cart(s, 1, ph, v) })
	return phi, theta
}

func presses(s int, start float64, delta int) []float64 {
	base := 0.0
	if start < 0 {
		base = -half(s)
	}
	vs := []float64{base, start}
	sign := 1
	if delta < 0 {
		sign = -1
	}
	v := start
	for n := delta; n != 0; n -= sign {
		v = Signed(s, Turn(s, v)+float64(sign))
		vs = append(vs, v)
	}
	return vs
}

func (f Frame) path(s int, start float64, delta int, at func(turn float64) Vec3) Path {
	vs := presses(s, start, delta)
	turn := float64(TurnTicks * s)
	p := Path{Pts: []Vec3{at(Turn(s, vs[0]))}}
	part := 0.0
	for i := 1; i < len(vs); i++ {
		a, b := Turn(s, vs[i-1]), Turn(s, vs[i])
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
