package Links

import (
	"fmt"
	"math"
	"slices"
)

func (p *Peer) setStart(m SetStart) []Envelope {
	l, ok := p.out[m.To]
	if !ok {
		return nil
	}
	l.SetStart(p.S, m.K, m.V)
	p.touched(m.To)
	return p.send(l)
}

func (p *Peer) press(m Press) []Envelope {
	l, ok := p.out[m.To]
	if !ok {
		return nil
	}
	lock := l.Lock.at(m.K)
	switch {
	case *lock == m.Sign:
		*lock = 0
	case l.Locking:
		*lock = m.Sign
	default:
		l.Step(p.S, m.K, m.Sign)
		p.touched(m.To)
		return p.send(l)
	}
	p.touched(m.To)
	return nil
}

func (p *Peer) setIncrement(m SetIncrement) {
	if l, ok := p.out[m.To]; ok && whole(m.V) {
		*l.Increment.at(m.K) = math.Round(m.V)
		p.touched(m.To)
	}
}

func (p *Peer) setLocking(m SetLocking) {
	if l, ok := p.out[m.To]; ok {
		l.Locking = m.On
		p.touched(m.To)
	}
}

func (p *Peer) pickReq(m PickReq) []Envelope {
	for from, on := range p.inPicked {
		if on && from != m.From {
			p.Changed.Refused = fmt.Sprintf("%s→%s: %s already has a picked link into it, from %s", m.From, p.ID, p.ID, from)
			return nil
		}
	}
	p.inPicked[m.From] = true
	return []Envelope{{To: m.From, Msg: PickSet{From: p.ID, On: true}}}
}

func (p *Peer) pickSet(m PickSet) {
	if l, ok := p.out[m.From]; ok {
		l.Picked = m.On
		p.touched(m.From)
	}
}

func (p *Peer) unpick(m Unpick) []Envelope {
	l, ok := p.out[m.To]
	if !ok || !l.Picked {
		return nil
	}
	l.Picked = false
	p.touched(m.To)
	return []Envelope{{To: m.To, Msg: PickDropped{From: p.ID}}}
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
	if m.V < 1 || m.V == p.S {
		return nil
	}
	k := float64(m.V) / float64(p.S)
	p.S = m.V
	p.Changed.S = true
	for _, to := range p.order {
		p.out[to].ScaleAngles(k)
		p.touched(to)
	}
	return []Envelope{{To: All, Msg: Scaled{ID: p.ID, S: p.S}}}
}

func (p *Peer) connect(m Connect) {
	if _, ok := p.out[m.To]; ok || m.To == p.ID {
		return
	}
	if _, ok := p.places[m.To]; !ok {
		return
	}
	l := NewLink(p.ID, m.To)
	l.Start = Measure(p.Frame, p.S, l.Dir(p.S), p.Pos(), p.posOf(m.To))
	p.out[m.To] = l
	p.order = append(p.order, m.To)
	p.touched(m.To)
}

func (p *Peer) drop(m Drop) []Envelope {
	l, ok := p.out[m.To]
	if !ok {
		return nil
	}
	delete(p.out, m.To)
	p.order = slices.DeleteFunc(p.order, func(t string) bool { return t == m.To })
	p.Changed.Dropped = append(p.Changed.Dropped, m.To)
	if l.Picked {
		return []Envelope{{To: m.To, Msg: PickDropped{From: p.ID}}}
	}
	return nil
}
