package Links

import (
	"maps"
	"slices"
)

func (p *Peer) send(l *Link) []Envelope {
	return []Envelope{{To: l.To, Msg: Put{From: p.ID, P: p.Pos(), D: p.vecOf(l), Keep: l.Name()}}}
}

func way(l *Link, all Lock, k string) int {
	if v := *l.Lock.at(k); v != 0 {
		return v
	}
	if v := *all.at(k); v != 0 {
		return v
	}
	return 1
}

func (p *Peer) stepAll(m StepAll) []Envelope {
	var out []Envelope
	for _, l := range p.picked() {
		if m.K == KBoth {
			l.Step(p.S, KPhi, m.Sign*way(l, m.LockAll, KPhi))
			l.Step(p.S, KTheta, m.Sign*way(l, m.LockAll, KTheta))
		} else {
			l.Step(p.S, m.K, m.Sign)
		}
		p.touched(l.To)
		out = append(out, Envelope{To: l.To, Msg: Put{From: p.ID, P: p.Pos(), D: p.vecOf(l), Round: m.Round, Step: true}})
	}
	p.round = m.Round
	held := p.held
	p.held = nil
	for _, h := range held {
		out = append(out, p.Handle(h)...)
	}
	return out
}

func (p *Peer) sat(keep string, keepPicked bool) []Envelope {
	p.places[p.ID] = p.At
	p.Changed.Place = true
	for _, to := range p.order {
		l := p.out[to]
		if l.Name() == keep || (keepPicked && l.Picked) {
			continue
		}
		l.Remeasure(p.Frame, p.S, p.Pos(), p.posOf(to))
		p.touched(to)
	}
	var out []Envelope
	for _, from := range slices.Sorted(maps.Keys(p.in)) {
		out = append(out, Envelope{To: from, Msg: Moved{ID: p.ID, At: p.At, Keep: keep, KeepPicked: keepPicked}})
	}
	return out
}

func (p *Peer) moved(m Moved) {
	p.places[m.ID] = m.At
	if l, ok := p.out[m.ID]; ok && l.Name() != m.Keep && !(m.KeepPicked && l.Picked) {
		l.Remeasure(p.Frame, p.S, p.Pos(), p.posOf(m.ID))
		p.touched(m.ID)
	}
}

func (p *Peer) take(m Put) []Envelope {
	p.At = p.Frame.GridAt(m.P.Add(m.D))
	return p.sat(m.Keep, m.Step)
}
