package Links

func (p *Peer) stepAll(m StepAll) []Envelope {
	shifted := p.shift()
	p.round, p.point, p.placed = m.Round, p.Pos(), p.pickedIn() == nil
	out := shifted
	if p.placed {
		out = append(append(out, p.tell()...), p.movedTo(true, false, Vec3{})...)
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
	p.places[m.From] = m.P
	p.point = p.Pos()
	return append(p.tell(), p.movedTo(true, false, Vec3{})...)
}

func (p *Peer) turnedBy(from string, d Vec3) []Envelope {
	out := p.sat(from, false, true, p.step(d))
	p.rebase()
	return out
}

func (p *Peer) step(d Vec3) Vec3 {
	was := p.Pos()
	*p.At = p.Frame.GridNear(was.Add(d), *p.At)
	return p.Pos().Sub(was)
}

func (p *Peer) sat(keep string, keepPicked, measure bool, by Vec3) []Envelope {
	p.Changed.Place = true
	for _, from := range p.order {
		if measure && from != keep && !(keepPicked && p.own[from].Picked) {
			p.remeasure(from, p.gapOf(p.own[from], p.refs[from]).Add(by))
		}
	}
	return p.movedTo(keepPicked, measure, by)
}

func (p *Peer) movedTo(keepPicked, measure bool, by Vec3) []Envelope {
	var out []Envelope
	for _, id := range p.neighbours() {
		out = append(out, Envelope{To: id, Msg: Moved{ID: p.ID, By: by, At: p.Pos(), KeepPicked: keepPicked, Measure: measure, Round: p.round}})
	}
	return out
}

func (p *Peer) moved(m Moved) []Envelope {
	p.heard[m.ID] = max(p.heard[m.ID], m.Round)
	p.places[m.ID] = m.At
	l, ok := p.own[m.ID]
	if !ok {
		return nil
	}
	if l.Picked && !(m.Measure && !m.KeepPicked) {
		return p.moveBy(m.By)
	}
	p.remeasure(m.ID, p.gapOf(l, p.refs[m.ID]).Sub(m.By))
	p.rebase()
	return nil
}

func (p *Peer) shift() []Envelope {
	var d Vec3
	for _, from := range p.order {
		v := p.vecOf(p.own[from])
		if old, ok := p.drawn[from]; ok && p.own[from].Picked {
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
	by := p.step(d)
	for _, from := range p.order {
		if l := p.own[from]; !l.Picked {
			p.remeasure(from, p.gapOf(l, p.refs[from]).Add(by))
		}
	}
	p.point = p.Pos()
	p.Changed.Place = true
	return p.movedTo(false, false, by)
}
