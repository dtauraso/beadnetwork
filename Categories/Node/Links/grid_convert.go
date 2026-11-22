package Links

func FreshCol(step int, live []Link, at Place, ref Ref) Col {
	c := Col{Step: step, At: at, Ref: ref}
	for _, l := range live {
		l.Picked = false
		l.Delta = Ints{S: l.Delta.S}
		c.In = append(c.In, InCell{Link: l})
	}
	return c
}

func (p *Peer) remap(steps, from []int) {
	old := p.grid
	var cols []Col
	for i, n := range steps {
		src := n
		if from != nil {
			src = from[i]
		}
		if c := old.at(src); src != 0 && c != nil {
			nc := c.clone()
			nc.Step = n
			cols = append(cols, nc)
			continue
		}
		cols = append(cols, FreshCol(n, p.Links(), *p.At, *p.Ref))
	}
	p.grid.Cols = cols
	p.sent = map[sentKey]OutLink{}
}

func (g Grid) Steps() []int {
	out := make([]int, len(g.Cols))
	for i, c := range g.Cols {
		out[i] = c.Step
	}
	return out
}
