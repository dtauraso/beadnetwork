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
	if m.K == KR {
		if l.Start.R+l.DeltaR+float64(m.Sign)*l.IncR < 0 {
			p.Changed.Refused = fmt.Sprintf("%s→%s Δr: r + Δr would be below 0, so Δr stays at %v", m.From, p.ID, l.DeltaR)
			return nil
		}
		l.Step(p.S, m.K, m.Sign)
		p.touched(m.From)
		return p.take(l)
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
	l, ok := p.own[m.From]
	switch {
	case !ok:
		return
	case m.K == KR && m.V > 0:
		l.IncR = m.V
	case m.K != KR && whole(m.V):
		*l.Increment.at(m.K) = math.Round(m.V)
	default:
		return
	}
	p.touched(m.From)
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
	if m.V < 1 {
		return nil
	}
	v, why := p.closestS(m.V)
	p.SNote = why
	p.Changed.S = true
	if v == p.S {
		return nil
	}
	k := float64(v) / float64(p.S)
	p.S = v
	for _, from := range p.order {
		p.own[from].ScaleAngles(k)
		p.touched(from)
	}
	return nil
}

func (p *Peer) closestS(v int) (int, string) {
	step, by := 1, ""
	for _, from := range p.order {
		d := p.own[from].Delta
		if n := p.S / gcd(p.S, gcd(int(math.Round(d.Phi)), int(math.Round(d.Theta)))); n > step {
			step, by = n, fmt.Sprintf("%s→%s's Δ (%g, %g)", from, p.ID, d.Phi, d.Theta)
		}
	}
	c := max(step, (2*v+step)/(2*step)*step)
	if c == v {
		return v, ""
	}
	return c, fmt.Sprintf("s %d would put %s between ticks; it stays whole only for s a multiple of %d, so s is %d", v, by, step, c)
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return max(a, -a)
}

func (p *Peer) connect(m Connect) []Envelope {
	if _, ok := p.own[m.From]; ok || m.From == p.ID {
		return nil
	}
	p.places[m.From] = m.At
	p.heard[m.From] = p.round
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
