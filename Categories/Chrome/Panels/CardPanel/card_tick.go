package CardPanel

func HasTick(f Field) bool {
	return f.Vector == VecStart && (f.Comp == CompPhi || f.Comp == CompTheta)
}

func TickOf(f Field) Field {
	return Field{Node: f.Node, Vector: VecTick, J: f.J, Comp: f.Comp}
}

func TickFields(node int) []Field {
	var out []Field
	for _, f := range NodeFields(node) {
		if HasTick(f) {
			out = append(out, TickOf(f))
		}
	}
	return out
}

func StateFields(node int) []Field {
	return append(NodeFields(node), TickFields(node)...)
}

func Step(start, tick, s, delta int) (int, int) {
	total := start*s + tick + delta
	q, r := total/s, total%s
	if r < 0 {
		q, r = q-1, r+s
	}
	return q, r
}

func (c *Card) ClampTicks(s int) bool {
	cleared := false
	for i := range c.Tick {
		v := &c.Tick[i]
		if v.Phi < 0 || v.Phi >= s || v.Theta < 0 || v.Theta >= s {
			*v, cleared = Vec{}, true
		}
	}
	return cleared
}
