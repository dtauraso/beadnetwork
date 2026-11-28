package Links

import (
	"maps"
	"slices"
)

type InCell struct {
	Link Link
	Sum  Angles
}

type OutCell struct {
	Link       Link
	Sum        Angles
	Start, End Vec3
	Steps      int
}

type Node struct {
	ID         string
	Ref        Ref
	At         Place
	Pos, Label Vec3
	In         []InCell
	Out        []OutCell
}

type NodeGeom struct {
	NodeR     float64
	LineSteps func(dist float64) int
}

type OutLink struct {
	Link Link
	At   Place
}

func (n Node) Links() []Link {
	out := make([]Link, len(n.In))
	for i, in := range n.In {
		out[i] = in.Link
	}
	return out
}

func (n Node) InFrom(from string) (InCell, bool) {
	for _, in := range n.In {
		if in.Link.From == from {
			return in, true
		}
	}
	return InCell{}, false
}

func (n Node) Clone() Node {
	n.In, n.Out = slices.Clone(n.In), slices.Clone(n.Out)
	return n
}

func (p *Peer) SetGeom(g NodeGeom) { p.geom = g }

func (p *Peer) Round() int { return p.round }

func (p *Peer) Node() Node { return p.node.Clone() }

func (p *Peer) Snap() []Envelope {
	n := Node{ID: p.ID, Ref: *p.Ref, At: *p.At, Pos: p.Pos()}
	n.Label = n.Pos.Add(Vec3{Y: p.geom.NodeR})
	var out []Envelope
	for _, from := range p.order {
		l := p.own[from]
		if l.Start.S != p.S {
			l.SetS(p.S)
		}
		n.In = append(n.In, InCell{Link: *l, Sum: p.Frame.Sum(*l, p.refs[from])})
		m := OutLink{Link: *l, At: *p.At}
		if was, ok := p.sent[from]; ok && was == m {
			continue
		}
		p.sent[from] = m
		out = append(out, Envelope{To: from, Msg: m})
	}
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		o, ok := p.outs[to]
		if !ok {
			o.Link = Link{From: p.ID, To: to}
		}
		o.Start = n.Pos
		o.Sum = p.Frame.Sum(o.Link, RefTurn(*p.Ref))
		if p.geom.LineSteps != nil {
			o.Steps = p.geom.LineSteps(o.End.Sub(o.Start).Length())
		}
		n.Out = append(n.Out, o)
	}
	p.node = n
	return out
}

func (p *Peer) outLink(m OutLink) {
	o := p.outs[m.Link.To]
	o.Link, o.End = m.Link, p.Frame.At(m.At)
	p.outs[m.Link.To] = o
}
