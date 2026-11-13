package Links

import "fmt"

func (p *Peer) pickReq(m PickReq) []Envelope {
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	if on := p.pickedIn(); on != nil && on.From != m.From {
		p.Changed.Refused = fmt.Sprintf("%s→%s: %s already has a picked link into it, from %s", m.From, p.ID, p.ID, on.From)
		return nil
	}
	l.Picked = true
	p.touched(m.From)
	return []Envelope{{To: m.From, Msg: PickSet{From: p.ID, On: true}}}
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

func (p *Peer) setIdx(m SetIdx) []Envelope {
	switch m.K {
	case KR:
		p.At.R = m.V
	case KTheta:
		p.At.Theta = m.V
	default:
		p.At.Phi = m.V
	}
	return p.sat("", false, true)
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
	p.places[m.From] = m.At
	p.heard[m.From] = p.round
	l := NewLink(m.From, p.ID)
	l.Order = NewLinkOrder
	l.Start = Measure(p.Frame, l.Dir(), p.posOf(m.From), p.Pos())
	l.SetS(p.S)
	p.own[m.From] = l
	if c := p.grid.at(p.grid.Shown); c != nil {
		c.In = append(c.In, InCell{Link: *l})
		p.bind()
	}
	p.drawn[m.From] = p.vecOf(l)
	p.order = append(p.order, m.From)
	p.touched(m.From)
	return []Envelope{{To: m.From, Msg: Linked{From: p.ID, At: *p.At}}}
}
