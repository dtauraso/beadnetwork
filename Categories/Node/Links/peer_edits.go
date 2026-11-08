package Links

import (
	"fmt"
	"maps"
	"slices"
)

func (p *Peer) setLink(m SetLink) []Envelope {
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	l.Start, l.Delta, l.Increment = m.Start, m.Delta, m.Increment
	p.touched(m.From)
	return p.take(l)
}

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
	return p.sat("", false)
}

func (p *Peer) setS(m SetS) []Envelope {
	if m.V < 1 {
		return nil
	}
	p.SNote = m.Note
	p.Changed.S = true
	if m.V == p.S {
		return nil
	}
	p.S = m.V
	p.scales[p.ID] = m.V
	var out []Envelope
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		out = append(out, Envelope{To: to, Msg: SSet{From: p.ID, S: m.V}})
	}
	return out
}

func (p *Peer) sSet(m SSet) {
	p.scales[m.From] = m.S
	if _, ok := p.own[m.From]; ok {
		p.touched(m.From)
	}
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
	p.own[m.From] = l
	p.order = append(p.order, m.From)
	p.touched(m.From)
	return []Envelope{{To: m.From, Msg: Linked{From: p.ID, At: p.At}}}
}
