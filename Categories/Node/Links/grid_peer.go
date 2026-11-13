package Links

import (
	"cmp"
	"maps"
	"slices"
)

type ShowStep struct {
	Level string
	Step  int
	Steps []int
	From  []int
}

type OutLink struct {
	Level string
	Step  int
	Link  Link
	At    Place
}

type WantOut struct {
	Level string
	From  string
}

type sentKey struct {
	Step int
	From string
}

func (p *Peer) SetGrid(g Grid, geom GridGeom) {
	p.grid, p.geom, p.sent = g, geom, map[sentKey]OutLink{}
	if len(p.grid.Cols) == 0 {
		p.grid.Cols = []Col{freshCol(1, p.Links(), *p.At, *p.Ref)}
	}
	if p.grid.at(p.grid.Shown) == nil {
		p.grid.Shown = p.grid.Cols[0].Step
	}
	for i := range p.grid.Cols {
		p.fitLinks(&p.grid.Cols[i])
	}
	p.fitBase()
	ref := *p.Ref
	p.bind()
	*p.Ref = ref
	p.fresh = true
}

func (p *Peer) bind() {
	c := p.grid.at(p.grid.Shown)
	if c == nil {
		return
	}
	for i := range c.In {
		p.own[c.In[i].Link.From] = &c.In[i].Link
	}
	p.At, p.Ref = &c.At, &c.Ref
}

func (p *Peer) fitLinks(c *Col) {
	var in []InCell
	for _, from := range p.order {
		cell, ok := c.InFrom(from)
		if !ok {
			cell.Link = *p.own[from]
			cell.Link.Picked = false
		}
		in = append(in, cell)
	}
	var out []OutCell
	for _, o := range c.Out {
		if p.out[o.Link.To] {
			out = append(out, o)
		}
	}
	c.In, c.Out = in, out
}

func (p *Peer) Grid() Grid { return p.grid.Clone() }

func (p *Peer) Level() string { return p.grid.Level }

func (p *Peer) Round() int { return p.round }

func (p *Peer) Snap() []Envelope {
	c := p.grid.at(p.grid.Shown)
	if c == nil {
		p.SetGrid(p.grid, p.geom)
		c = p.grid.at(p.grid.Shown)
	}
	for to := range p.out {
		c.outTo(to).End = p.posOf(to)
	}
	c.Out = slices.DeleteFunc(c.Out, func(o OutCell) bool { return !p.out[o.Link.To] })
	slices.SortFunc(c.Out, func(a, b OutCell) int { return cmp.Compare(a.Link.To, b.Link.To) })
	p.fitTotals()
	for i := range p.grid.Cols {
		p.compute(&p.grid.Cols[i])
	}
	return p.outLinks()
}

func (p *Peer) compute(c *Col) {
	c.Ref = *p.Ref
	c.Pos = p.Frame.At(c.At)
	c.Label = c.Pos.Add(Vec3{Y: p.geom.NodeR})
	for i := range c.In {
		l := &c.In[i].Link
		if l.Start.S != p.S {
			l.SetS(p.S)
		}
		c.In[i].Sum = p.Frame.Sum(*l, p.refs[l.From])
	}
	for i := range c.Out {
		o := &c.Out[i]
		o.Start = c.Pos
		o.Sum = p.Frame.Sum(o.Link, RefTurn(*p.Ref))
		if p.geom.LineSteps != nil {
			o.Steps = p.geom.LineSteps(o.End.Sub(o.Start).Length())
		}
	}
}

func (p *Peer) outLinks() []Envelope {
	var out []Envelope
	for _, c := range p.grid.Cols {
		for _, in := range c.In {
			m := OutLink{Level: p.grid.Level, Step: c.Step, Link: in.Link, At: c.At}
			k := sentKey{c.Step, in.Link.From}
			if p.sent[k] == m {
				continue
			}
			p.sent[k] = m
			out = append(out, Envelope{To: in.Link.From, Msg: m})
		}
	}
	return out
}

func (p *Peer) outLink(m OutLink) {
	if m.Level != p.grid.Level {
		return
	}
	if c := p.grid.at(m.Step); c != nil {
		o := c.outTo(m.Link.To)
		o.Link, o.End = m.Link, p.Frame.At(m.At)
	}
}

func (p *Peer) wantOut(m WantOut) {
	if m.Level != p.grid.Level {
		return
	}
	for k := range p.sent {
		if k.From == m.From {
			delete(p.sent, k)
		}
	}
}

func (p *Peer) showStep(m ShowStep) []Envelope {
	var out []Envelope
	fresh := p.fresh
	p.fresh = false
	if m.Level == p.grid.Level && !fresh {
		out = p.Snap()
	}
	was := p.grid.ShownCol()
	at, ref := *p.At, *p.Ref
	walk := m.From == nil && m.Step != was.Step
	if m.Steps != nil {
		p.remap(m.Steps, m.From)
		p.bind()
	}
	c := p.grid.at(m.Step)
	if c == nil {
		return out
	}
	if walk {
		for from, b := range p.grid.Base {
			if in, ok := was.InFrom(from); ok {
				p.grid.Base[from] = b.Add(in.Link.Delta)
			}
		}
	}
	p.grid.Shown = m.Step
	p.bind()
	if walk {
		*p.At = at
	}
	*p.Ref = ref
	p.fitTotals()
	p.Changed.Place = true
	for _, from := range p.order {
		p.touched(from)
		out = append(out, Envelope{To: from, Msg: PickSet{From: p.ID, On: p.own[from].Picked}})
	}
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		out = append(out, Envelope{To: to, Msg: WantOut{Level: p.grid.Level, From: p.ID}})
	}
	return out
}
