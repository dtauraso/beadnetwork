package Links

import (
	"fmt"
	"math"
	"slices"
)

func (p *Peer) setStart(m SetStart) []Envelope {
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	l.SetStart(p.S, m.K, m.V)
	p.touched(m.From)
	return p.take(l)
}

func (p *Peer) press(m Press) []Envelope {
	l, ok := p.own[m.From]
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
		p.touched(m.From)
		return p.take(l)
	}
	p.touched(m.From)
	return nil
}

func (p *Peer) setIncrement(m SetIncrement) {
	if l, ok := p.own[m.From]; ok && whole(m.V) {
		*l.Increment.at(m.K) = math.Round(m.V)
		p.touched(m.From)
	}
}

func (p *Peer) setLocking(m SetLocking) {
	if l, ok := p.own[m.From]; ok {
		l.Locking = m.On
		p.touched(m.From)
	}
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
	if m.V < 1 || m.V == p.S {
		return nil
	}
	k := float64(m.V) / float64(p.S)
	p.S = m.V
	p.Changed.S = true
	for _, from := range p.order {
		p.own[from].ScaleAngles(k)
		p.touched(from)
	}
	return nil
}

func (p *Peer) connect(m Connect) []Envelope {
	if _, ok := p.own[m.From]; ok || m.From == p.ID {
		return nil
	}
	p.places[m.From] = m.At
	l := NewLink(m.From, p.ID)
	l.Order = NewLinkOrder
	l.Start = Measure(p.Frame, p.S, l.Dir(p.S), p.posOf(m.From), p.Pos())
	p.own[m.From] = l
	p.order = append(p.order, m.From)
	p.touched(m.From)
	return []Envelope{{To: m.From, Msg: Linked{From: p.ID, At: p.At}}}
}

func (p *Peer) drop(m Drop) []Envelope {
	if _, ok := p.own[m.From]; !ok {
		return nil
	}
	delete(p.own, m.From)
	p.order = slices.DeleteFunc(p.order, func(f string) bool { return f == m.From })
	p.Changed.Dropped = append(p.Changed.Dropped, m.From)
	return []Envelope{{To: m.From, Msg: Unlinked{From: p.ID}}}
}
