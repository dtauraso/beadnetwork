package Links

import (
	"fmt"
	"maps"
	"slices"
)

func (p *Peer) HasPickedIn() bool { return p.pickedIn() != nil }

func (p *Peer) PlayIn(from string, frac float64) (Vec3, bool) {
	l, ok := p.own[from]
	if !ok || !l.Picked {
		return Vec3{}, false
	}
	return p.Frame.LinkVecAt(*l, p.refs[from], frac), true
}

func (p *Peer) PlayOutVecs(frac float64) (vecs []Vec3, picked []bool) {
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		v := Vec3{}
		if p.outPicked[to] {
			v = p.outVecAt(to, frac)
		}
		vecs = append(vecs, v)
		picked = append(picked, p.outPicked[to])
	}
	return vecs, picked
}

func (p *Peer) outVecAt(to string, frac float64) Vec3 {
	if c := p.grid.at(p.grid.Shown); c != nil {
		for _, o := range c.Out {
			if o.Link.To == to {
				return p.Frame.LinkVecAt(o.Link, RefTurn(*p.Ref), frac)
			}
		}
	}
	return Vec3{}
}

func (p *Peer) PlayTo(run int, at Vec3, k, n int) []Envelope {
	var out []Envelope
	for _, to := range slices.Sorted(maps.Keys(p.outPicked)) {
		if p.outPicked[to] {
			out = append(out, Envelope{To: to, Msg: PlayAt{Run: run, From: p.ID, P: at, K: k, N: n}})
		}
	}
	return out
}

func (p *Peer) pickReq(m PickReq) []Envelope {
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	var out []Envelope
	if on := p.pickedIn(); on != nil && on.From != m.From {
		p.Changed.Refused = fmt.Sprintf("moved the pick from %s→%s to %s→%s: one picked link into a node at a time", on.From, p.ID, m.From, p.ID)
		out = p.unpick(Unpick{From: on.From})
	}
	l.Picked = true
	p.touched(m.From)
	return append(out, Envelope{To: m.From, Msg: PickSet{From: p.ID, On: true}})
}

func (p *Peer) pickSet(m PickSet) []Envelope {
	was := p.outPicked[m.From]
	p.outPicked[m.From] = m.On
	if m.On && !was && p.placed {
		return []Envelope{{To: m.From, Msg: Put{From: p.ID, P: p.point, Round: p.round}}}
	}
	return nil
}

func (p *Peer) unpick(m Unpick) []Envelope {
	l, ok := p.own[m.From]
	if !ok || !l.Picked {
		return nil
	}
	l.Picked = false
	p.touched(m.From)
	return []Envelope{{To: m.From, Msg: PickSet{From: p.ID, On: false}}}
}

func (p *Peer) setS(m SetS) []Envelope {
	if m.V < 1 {
		return nil
	}
	p.Changed.S = true
	if p.Ref.S != m.V {
		p.Ref.S = m.V
		p.Changed.Ref = true
	}
	if m.V == p.S {
		return nil
	}
	p.S = m.V
	for _, from := range p.order {
		p.own[from].SetS(m.V)
		p.touched(from)
	}
	return nil
}

func (p *Peer) connect(m Connect) []Envelope {
	if _, ok := p.own[m.From]; ok || m.From == p.ID {
		return nil
	}
	p.heard[m.From] = p.round
	l := NewLink(m.From, p.ID)
	l.Order = NewLinkOrder
	l.Start = Measure(p.Frame, l.Dir(), p.Frame.At(m.At), p.Pos())
	l.SetS(p.S)
	p.own[m.From] = l
	if c := p.grid.at(p.grid.Shown); c != nil {
		c.In = append(c.In, InCell{Link: *l})
		p.bind()
	}
	p.drawn[m.From] = p.vecOf(l)
	p.order = append(p.order, m.From)
	p.touched(m.From)
	return []Envelope{{To: m.From, Msg: Linked{From: p.ID}}}
}
