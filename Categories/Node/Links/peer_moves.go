package Links

import "slices"

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
	for _, l := range p.picked() {
		if m.K == KBoth {
			l.Step(p.S, KPhi, m.Sign*way(l, m.LockAll, KPhi))
			l.Step(p.S, KTheta, m.Sign*way(l, m.LockAll, KTheta))
		} else {
			l.Step(p.S, m.K, m.Sign)
		}
		p.touched(l.To)
	}
	p.round, p.roundAt = m.Round, p.Pos()
	var out []Envelope
	if p.hasPickedIn() {
		for _, l := range p.picked() {
			out = append(out, Envelope{To: l.To, Msg: Trail{Round: m.Round, Hops: []Hop{{ID: p.ID, P: p.roundAt, D: p.vecOf(l)}}}})
		}
	} else {
		out = p.cascade(m.Round, nil)
	}
	held := p.held
	p.held = nil
	for _, h := range held {
		out = append(out, p.Handle(h)...)
	}
	return out
}

func (p *Peer) cascade(round int, skip map[string]bool) []Envelope {
	var out []Envelope
	for _, l := range p.picked() {
		if skip[l.To] {
			continue
		}
		out = append(out, Envelope{To: l.To, Msg: Put{From: p.ID, P: p.Pos(), D: p.vecOf(l), Round: round, Cascade: true}})
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
	return []Envelope{{To: All, Msg: Moved{ID: p.ID, At: p.At, S: p.S, Keep: keep, KeepPicked: keepPicked}}}
}

func (p *Peer) moved(m Moved) {
	p.places[m.ID] = m.At
	if m.S > 0 {
		p.scales[m.ID] = m.S
	}
	if l, ok := p.out[m.ID]; ok && l.Name() != m.Keep && !(m.KeepPicked && l.Picked) {
		l.Remeasure(p.Frame, p.S, p.Pos(), p.posOf(m.ID))
		p.touched(m.ID)
	}
}

func (p *Peer) take(m Put) []Envelope {
	p.At = p.Frame.GridAt(m.P.Add(m.D))
	out := p.sat(m.Keep, m.Cascade)
	if m.Cascade {
		out = append(out, p.cascade(m.Round, nil)...)
	}
	return out
}

func (p *Peer) trail(m Trail) []Envelope {
	if m.Hops[0].ID == p.ID {
		return p.ring(m)
	}
	for _, h := range m.Hops {
		if h.ID == p.ID {
			return nil
		}
	}
	var out []Envelope
	for _, l := range p.picked() {
		hops := append(slices.Clone(m.Hops), Hop{ID: p.ID, P: p.roundAt, D: p.vecOf(l)})
		out = append(out, Envelope{To: l.To, Msg: Trail{Round: m.Round, Hops: hops}})
	}
	return out
}

func (p *Peer) ring(m Trail) []Envelope {
	n := float64(len(m.Hops))
	var gap, mid Vec3
	for _, h := range m.Hops {
		gap = gap.Add(h.D)
		mid = mid.Add(h.P)
	}
	gap, mid = gap.Scale(1/n), mid.Scale(1/n)
	qs := make([]Vec3, len(m.Hops))
	var qmid Vec3
	for i := 1; i < len(m.Hops); i++ {
		qs[i] = qs[i-1].Add(m.Hops[i-1].D.Sub(gap))
	}
	for _, q := range qs {
		qmid = qmid.Add(q)
	}
	shift := mid.Sub(qmid.Scale(1 / n))
	p.At = p.Frame.GridAt(qs[0].Add(shift))
	out := p.sat("", true)
	inRing := map[string]bool{}
	for _, h := range m.Hops {
		inRing[h.ID] = true
	}
	return append(out, p.cascade(m.Round, inRing)...)
}
