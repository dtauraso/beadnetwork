package Links

import "slices"

type InCell struct {
	Link Link
	Sum  Start
}

type OutCell struct {
	Link       Link
	Sum        Start
	Start, End Vec3
	Steps      int
}

type Col struct {
	Step       int
	Ref        Ref
	At         Place
	Pos, Label Vec3
	In         []InCell
	Out        []OutCell
}

type Grid struct {
	ID    string
	Level string
	Shown int
	Cols  []Col
}

type GridGeom struct {
	NodeR     float64
	LineSteps func(dist float64) int
}

func (g Grid) Steps() []int {
	out := make([]int, len(g.Cols))
	for i, c := range g.Cols {
		out[i] = c.Step
	}
	return out
}

func (g Grid) Col(step int) (Col, bool) {
	if c := g.at(step); c != nil {
		return *c, true
	}
	return Col{}, false
}

func (g Grid) ShownCol() Col {
	c, _ := g.Col(g.Shown)
	return c
}

func (g *Grid) at(step int) *Col {
	for i := range g.Cols {
		if g.Cols[i].Step == step {
			return &g.Cols[i]
		}
	}
	return nil
}

func (c Col) clone() Col {
	c.In, c.Out = slices.Clone(c.In), slices.Clone(c.Out)
	return c
}

func (g Grid) Clone() Grid {
	cols := make([]Col, len(g.Cols))
	for i, c := range g.Cols {
		cols[i] = c.clone()
	}
	g.Cols = cols
	return g
}

func (c Col) Links() []Link {
	out := make([]Link, len(c.In))
	for i, in := range c.In {
		out[i] = in.Link
	}
	return out
}

func (c Col) InFrom(from string) (InCell, bool) {
	for _, in := range c.In {
		if in.Link.From == from {
			return in, true
		}
	}
	return InCell{}, false
}

func (c *Col) outTo(to string) *OutCell {
	for i := range c.Out {
		if c.Out[i].Link.To == to {
			return &c.Out[i]
		}
	}
	c.Out = append(c.Out, OutCell{Link: Link{To: to}})
	return &c.Out[len(c.Out)-1]
}

func (f Frame) Sum(l Link, ref Turn) Start {
	d, a := l.Dir(), ref.In(1)
	return Measure(f, Start{Phi: d.Phi + a.Phi, Theta: d.Theta + a.Theta}, Vec3{}, f.LinkVec(l, ref))
}

func Normalized(l Link, s int) Link {
	l.Start.Phi, l.Start.Theta = FromSigned(1, l.Start.Phi), FromSigned(1, l.Start.Theta)
	l.Start = l.Start.Pulled()
	l.SetS(s)
	return l
}
