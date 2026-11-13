package Links

import (
	"fmt"
	"slices"
)

func way(l *Link, all Lock, k string) int {
	if v := *l.Lock.at(k); v != 0 {
		return v
	}
	if v := *all.at(k); v != 0 {
		return v
	}
	return 1
}

func (p *Peer) stepR(l *Link, sign int) {
	if l.Start.R+float64(l.Sum().R+sign*l.Increment.R) < 0 {
		p.Changed.Refused = fmt.Sprintf("%s→%s Δr: r + Δr would be below 0, so Δr stays at %v", l.From, p.ID, l.Delta.R)
		return
	}
	l.Step(KR, sign)
}

func (p *Peer) stepAll(m StepAll) []Envelope {
	l := p.pickedIn()
	if m.From != "" {
		l = p.own[m.From]
	}
	if l != nil {
		fixed := ""
		for _, id := range []string{p.ID, l.From} {
			if fixed == "" && slices.Contains(m.FixedR, id) {
				fixed = id
			}
		}
		if fixed != "" && m.Sign != 0 && (m.K == KBoth || m.K == KR) {
			p.Changed.Refused = fmt.Sprintf("%s→%s Δr: %s has more than one incoming link, so no radius it is part of can change; Δr stays at %v", l.From, p.ID, fixed, l.Delta.R)
		}
		switch m.K {
		case KBoth:
			l.Step(KPhi, m.Sign*way(l, m.LockAll, KPhi))
			l.Step(KTheta, m.Sign*way(l, m.LockAll, KTheta))
			if fixed == "" {
				p.stepR(l, m.Sign*way(l, m.LockAll, KR))
			}
		case KR:
			if fixed == "" {
				p.stepR(l, m.Sign)
			}
		default:
			l.Step(m.K, m.Sign)
		}
		p.touched(l.From)
	}
	shifted := p.shift()
	p.round, p.point, p.placed = m.Round, p.Pos(), p.pickedIn() == nil
	out := shifted
	if p.placed {
		out = append(append(out, p.tell()...), p.movedTo(true, false)...)
	}
	held := p.held
	p.held = nil
	for _, h := range held {
		out = append(out, p.Handle(h)...)
	}
	return out
}

func (p *Peer) tell() []Envelope {
	var out []Envelope
	for _, to := range p.neighbours() {
		if p.outPicked[to] {
			out = append(out, Envelope{To: to, Msg: Put{From: p.ID, P: p.point, Round: p.round}})
		}
	}
	return out
}

func (p *Peer) recover(m Put) []Envelope {
	l, ok := p.own[m.From]
	if p.placed || !ok || !l.Picked {
		return nil
	}
	p.placed = true
	p.point = p.Pos()
	return append(p.tell(), p.movedTo(true, false)...)
}

func (p *Peer) take(l *Link, measure bool) []Envelope {
	*p.At = p.Frame.GridAt(p.posOf(l.From).Add(p.vecOf(l)))
	out := p.sat(l.From, false, measure)
	p.rebase()
	return out
}

func (p *Peer) sat(keep string, keepPicked, measure bool) []Envelope {
	p.Changed.Place = true
	for _, from := range p.order {
		if measure && from != keep && !(keepPicked && p.own[from].Picked) {
			p.remeasure(from)
		}
	}
	return p.movedTo(keepPicked, measure)
}

func (p *Peer) movedTo(keepPicked, measure bool) []Envelope {
	var out []Envelope
	for _, id := range p.neighbours() {
		out = append(out, Envelope{To: id, Msg: Moved{ID: p.ID, At: *p.At, KeepPicked: keepPicked, Measure: measure, Round: p.round}})
	}
	return out
}

func (p *Peer) moved(m Moved) []Envelope {
	was := p.posOf(m.ID)
	p.places[m.ID] = m.At
	p.heard[m.ID] = max(p.heard[m.ID], m.Round)
	l, ok := p.own[m.ID]
	switch {
	case !ok:
		return nil
	case m.Measure && !(m.KeepPicked && l.Picked):
		p.remeasure(m.ID)
		p.rebase()
	case l.Picked:
		return p.moveBy(p.posOf(m.ID).Sub(was))
	}
	return nil
}

func (p *Peer) shift() []Envelope {
	var d Vec3
	for _, from := range p.order {
		v := p.vecOf(p.own[from])
		if old, ok := p.drawn[from]; ok {
			d = d.Add(v.Sub(old))
		}
		p.drawn[from] = v
	}
	return p.moveBy(d)
}

func (p *Peer) rebase() {
	clear(p.drawn)
	p.shift()
}

func (p *Peer) moveBy(d Vec3) []Envelope {
	if d.Length() < 1e-12 {
		return nil
	}
	*p.At = p.Frame.GridAt(p.Pos().Add(d))
	p.point = p.Pos()
	p.Changed.Place = true
	return p.movedTo(false, false)
}

func (p *Peer) refit(walk bool) []Envelope {
	p.fitTotals()
	if walk {
		return p.shift()
	}
	p.rebase()
	return nil
}

func (p *Peer) count(was, now *Col) {
	for from, b := range p.grid.Base {
		if in, ok := was.InFrom(from); ok && !was.Counted {
			b = b.Add(in.Link.Delta)
		}
		if in, ok := now.InFrom(from); ok && now.Counted {
			b = b.Sub(in.Link.Delta)
		}
		p.grid.Base[from] = b
	}
	was.Counted, now.Counted = true, false
}

func (p *Peer) remeasure(from string) {
	l := p.own[from]
	l.Remeasure(p.Frame, p.refs[from], p.posOf(from), p.Pos())
	p.setOrigin(from, l.Start)
	p.touched(from)
}
