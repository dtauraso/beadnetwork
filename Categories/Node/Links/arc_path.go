package Links

const pathSteps = 8

func (f Frame) Arcs(s int, l Link) (phi, theta []Vec3) {
	ph := Turn(s, l.Dir(s).Phi)
	phi = f.path(s, l.Start.Phi, l.Delta.Phi, func(v float64) Vec3 { return f.cart(s, 1, v, 0) })
	theta = f.path(s, l.Start.Theta, l.Delta.Theta, func(v float64) Vec3 { return f.cart(s, 1, ph, v) })
	return phi, theta
}

func presses(s int, start float64, delta int) []float64 {
	left := start < 0
	base := 0.0
	if left {
		base = -half(s)
	}
	vs := []float64{base, start}
	sign := 1
	if delta < 0 {
		sign = -1
	}
	v := start
	for n := delta; n != 0; n -= sign {
		v = PressesIn(s, left, v, sign)
		vs = append(vs, v)
	}
	return vs
}

func (f Frame) path(s int, start float64, delta int, at func(turn float64) Vec3) []Vec3 {
	vs := presses(s, start, delta)
	turn := float64(TurnTicks * s)
	out := []Vec3{at(Turn(s, vs[0]))}
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
			out = append(out, at(Wrap(s, a+(b-a)*float64(j)/pathSteps)))
		}
		out = append(out, at(end))
	}
	return out
}
