package Links

import (
	"fmt"
	"math"
)

const (
	EditStart = "start"
	EditOne   = "one"
)

type StartEdit struct {
	From    string
	Op      string
	K       string
	Sign    int
	Typed   float64
	IsTyped bool
}

func (p *Peer) amount(m Amount) []Envelope {
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	l.Delta = l.Delta.Add(m.D)
	l.MarkWays()
	p.touched(m.From)
	l.Moving, l.Part = Ints{}, 0
	if !l.Picked {
		l.Start.Less(m.D)
		p.drawn[m.From] = p.vecOf(l)
		return nil
	}
	if m.Play {
		l.Moving = Ints{Phi: m.D.Phi, Theta: m.D.Theta, R: m.D.R}
	}
	return p.shift()
}

func (p *Peer) playPart(m PlayPart) []Envelope {
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	of := max(1, m.Of)
	l.Part = float64(min(m.Press, of)) / float64(of)
	if m.Press >= of {
		l.Moving, l.Part = Ints{}, 0
	}
	p.touched(m.From)
	return p.shift()
}

func (p *Peer) startEdit(m StartEdit) []Envelope {
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	name := fmt.Sprintf("start %s→%s", m.From, p.ID)
	if !l.Picked {
		p.Changed.Refused = fmt.Sprintf("%s: %s→%s is not picked, so its start does not move %s; the start stays", name, m.From, p.ID, p.ID)
		return nil
	}
	t := *l
	if refused := editStart(&t, m, name); refused != "" {
		p.Changed.Refused = refused
		return nil
	}
	was := p.vecOf(l)
	l.Start = t.Start
	p.touched(m.From)
	return p.turnedBy(m.From, p.vecOf(l).Sub(was))
}

func editStart(l *Link, m StartEdit, name string) string {
	sign, k := m.Sign, m.K
	if m.Op == EditOne {
		if sign == 0 {
			return ""
		}
		if k == KR && l.Dir().R+float64(l.OneCarry(k, sign)) < 0 {
			return fmt.Sprintf("%s 1r: r + Δr would be below 0, so 1r stays at %d", name, l.Start.One.R)
		}
		l.StepOne(k, sign)
		return ""
	}
	if k != KR {
		if m.IsTyped {
			l.Start.SetTick(k, int(math.Floor(m.Typed)))
		} else {
			l.Start.Turn(k, sign)
		}
		return ""
	}
	v := l.Start.R + float64(sign)
	if m.IsTyped {
		v = m.Typed
	}
	if v < 0 || v+float64(l.Delta.R) < 0 {
		return fmt.Sprintf("%s r %v: neither r nor r + Δr can be negative, so r stays at %v", name, v, l.Start.R)
	}
	l.Start.R = v
	return ""
}
