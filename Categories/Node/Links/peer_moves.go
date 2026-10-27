package Links

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
	if l := p.pickedIn(); l != nil {
		if m.K == KBoth {
			l.Step(p.S, KPhi, m.Sign*way(l, m.LockAll, KPhi))
			l.Step(p.S, KTheta, m.Sign*way(l, m.LockAll, KTheta))
		} else {
			l.Step(p.S, m.K, m.Sign)
		}
		p.touched(l.From)
	}
	p.round, p.point, p.placed = m.Round, p.Pos(), p.pickedIn() == nil
	var out []Envelope
	if p.placed {
		out = append(p.tell(), p.movedTo(true)...)
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
	p.At = p.Frame.GridAt(m.P.Add(p.vecOf(l)))
	p.point = p.Pos()
	return append(p.sat("", true), p.tell()...)
}

func (p *Peer) take(l *Link) []Envelope {
	p.At = p.Frame.GridAt(p.posOf(l.From).Add(p.vecOf(l)))
	return p.sat(l.From, false)
}

func (p *Peer) sat(keep string, keepPicked bool) []Envelope {
	p.places[p.ID] = p.At
	p.Changed.Place = true
	for _, from := range p.order {
		l := p.own[from]
		if from == keep || (keepPicked && l.Picked) {
			continue
		}
		l.Remeasure(p.Frame, p.S, p.posOf(from), p.Pos())
		p.touched(from)
	}
	return p.movedTo(keepPicked)
}

func (p *Peer) movedTo(keepPicked bool) []Envelope {
	var out []Envelope
	for _, id := range p.neighbours() {
		out = append(out, Envelope{To: id, Msg: Moved{ID: p.ID, At: p.At, KeepPicked: keepPicked, Round: p.round}})
	}
	return out
}

func (p *Peer) moved(m Moved) {
	p.places[m.ID] = m.At
	p.heard[m.ID] = max(p.heard[m.ID], m.Round)
	if l, ok := p.own[m.ID]; ok && !(m.KeepPicked && l.Picked) {
		l.Remeasure(p.Frame, p.S, p.posOf(m.ID), p.Pos())
		p.touched(m.ID)
	}
}
