package Links

import (
	"maps"
	"math"
	"slices"
)

type Turn struct{ Phi, Theta float64 }

func RefTurn(s int, ref Ticks) Turn {
	t := float64(TurnTicks * max(s, 1))
	return Turn{Phi: float64(ref.Phi) / t, Theta: float64(ref.Theta) / t}
}

func (r Turn) In(s int) Start {
	t := float64(TurnTicks * s)
	return Start{Phi: r.Phi * t, Theta: r.Theta * t}
}

func (l *Link) DirFrom(s int, ref Turn) Start {
	d, a := l.Dir(s), ref.In(s)
	return Start{Phi: Wrap(s, d.Phi+a.Phi), Theta: Wrap(s, d.Theta+a.Theta), R: d.R}
}

type SetRef struct {
	K string
	V int
}

type RefSet struct {
	From string
	Ref  Turn
}

func (p *Peer) setRef(m SetRef) []Envelope {
	if m.K == KTheta {
		p.Ref.Theta = m.V
	} else {
		p.Ref.Phi = m.V
	}
	p.Changed.Ref = true
	p.refs[p.ID] = RefTurn(p.S, p.Ref)
	var out []Envelope
	for _, to := range slices.Sorted(maps.Keys(p.out)) {
		out = append(out, Envelope{To: to, Msg: RefSet{From: p.ID, Ref: p.refs[p.ID]}})
	}
	return out
}

func (p *Peer) refSet(m RefSet) []Envelope {
	p.refs[m.From] = m.Ref
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	p.touched(m.From)
	return p.take(l)
}

func (p *Peer) RefOf(id string) Turn { return p.refs[id] }

func (p *Peer) RefVecs() (phi, theta Vec3) {
	a := p.refs[p.ID].In(p.S)
	return p.Frame.cart(p.S, 1, a.Phi, 0), p.Frame.cart(p.S, 1, half(p.S)/2, a.Theta)
}

func (p *Peer) scaleRef(k float64) {
	p.Ref.Phi = int(math.Round(float64(p.Ref.Phi) * k))
	p.Ref.Theta = int(math.Round(float64(p.Ref.Theta) * k))
	p.refs[p.ID] = RefTurn(p.S, p.Ref)
	p.Changed.Ref = true
}
