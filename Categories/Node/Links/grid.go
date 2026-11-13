package Links

import (
	"fmt"
	"maps"
	"slices"
)

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
	Base  map[string]Base
}

type Base struct {
	Origin Start
	Total  Ints
}

func (b Base) Plus(d Ints) Start {
	s := b.Origin
	s.Phi = Wrap(1, s.Phi+float64(b.Total.Phi+d.Phi))
	s.Theta = Wrap(1, s.Theta+float64(b.Total.Theta+d.Theta))
	s.R += float64(b.Total.R + d.R)
	return s.Pulled()
}

func endOf(l Link) Start { return Base{Origin: l.Start}.Plus(l.Delta) }

func baseOf(start Start, total Ints) Base {
	o := start
	o.Phi = Wrap(1, o.Phi-float64(total.Phi))
	o.Theta = Wrap(1, o.Theta-float64(total.Theta))
	o.R -= float64(total.R)
	return Base{Origin: o.Pulled(), Total: total}
}

func (p *Peer) fitBase() {
	if p.grid.Base == nil {
		p.grid.Base = map[string]Base{}
	}
	shown := p.grid.ShownCol()
	for from := range p.grid.Base {
		if _, ok := p.own[from]; !ok {
			delete(p.grid.Base, from)
		}
	}
	for _, from := range p.order {
		b, ok := p.grid.Base[from]
		if !ok {
			start := p.own[from].Start
			if in, ok := shown.InFrom(from); ok {
				start = in.Link.Start
			}
			b = baseOf(start, Ints{})
		}
		b.Origin.S = p.S
		p.grid.Base[from] = Base{Origin: b.Origin.Pulled(), Total: b.Total}
	}
}

func (b Base) Add(d Ints) Base {
	b.Total.Phi, b.Total.Theta, b.Total.R = b.Total.Phi+d.Phi, b.Total.Theta+d.Theta, b.Total.R+d.R
	return b
}

func (p *Peer) deriveStarts() {
	p.fitBase()
	shown := p.grid.ShownCol()
	for i := range p.grid.Cols {
		c := &p.grid.Cols[i]
		if c.Step == shown.Step {
			continue
		}
		for j := range c.In {
			if was, ok := shown.InFrom(c.In[j].Link.From); ok {
				c.In[j].Link.Start = endOf(was.Link)
				c.In[j].Link.SetS(p.S)
			}
		}
	}
}

func (p *Peer) originEdit(l *Link, m StepEdit) []Envelope {
	p.fitBase()
	b := p.grid.Base[m.From]
	t := *l
	t.Start = b.Origin
	t.SetS(p.S)
	if refused := editLink(&t, m, fmt.Sprintf("start %s→%s", m.From, p.ID)); refused != "" {
		p.Changed.Refused = refused
		return nil
	}
	dPhi, dTheta, dR := t.Start.Phi-b.Origin.Phi, t.Start.Theta-b.Origin.Theta, t.Start.R-b.Origin.R
	for i := range p.grid.Cols {
		c := &p.grid.Cols[i]
		for j := range c.In {
			if s := &c.In[j].Link.Start; c.In[j].Link.From == m.From {
				s.Phi, s.Theta, s.R, s.One = Wrap(1, s.Phi+dPhi), Wrap(1, s.Theta+dTheta), s.R+dR, t.Start.One
				*s = s.Pulled()
			}
		}
	}
	b.Origin = t.Start
	p.grid.Base[m.From] = b
	p.touched(m.From)
	return p.take(l)
}

type GridGeom struct {
	NodeR     float64
	LineSteps func(dist float64) int
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
	g.Base = maps.Clone(g.Base)
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
