package Links

import (
	"fmt"
	"math"
)

const (
	EditStart   = "start"
	EditOne     = "one"
	EditPress   = "press"
	EditInc     = "increment"
	EditPick    = "pick"
	EditHead    = "head"
	EditHeadInc = "head-increment"
	EditHeadOne = "head-one"
)

type StepEdit struct {
	Step    int
	From    string
	Op      string
	K       string
	Sign    int
	Typed   float64
	IsTyped bool
	Locks   Lock
	FixedR  bool
}

func (p *Peer) stepEdit(m StepEdit) []Envelope {
	l, ok := p.own[m.From]
	if !ok {
		return nil
	}
	name := fmt.Sprintf("step %d %s→%s", m.Step, m.From, p.ID)
	if m.Step != p.grid.Shown {
		p.Changed.Refused = fmt.Sprintf("%s: step %d is not the step shown (step %d is), so nothing changed", name, m.Step, p.grid.Shown)
		return nil
	}
	if m.Op == EditPick {
		if l.Picked {
			return p.unpick(Unpick{From: m.From})
		}
		return p.pickReq(PickReq{From: m.From})
	}
	if refused := editLink(l, m, name); refused != "" {
		p.Changed.Refused = refused
		return nil
	}
	p.touched(m.From)
	return p.take(l)
}

func editLink(l *Link, m StepEdit, name string) string {
	sign, k := m.Sign, m.K
	switch m.Op {
	case EditStart:
		v := map[string]float64{KPhi: l.Start.Phi, KTheta: l.Start.Theta, KR: l.Start.R}[k] + float64(sign)
		if m.IsTyped && k == KR {
			v = m.Typed
		} else if m.IsTyped {
			v = l.Start.FromTick(int(math.Floor(m.Typed)))
		}
		if k == KR && (v < 0 || v+float64(l.Delta.R) < 0) {
			return fmt.Sprintf("%s r %v: neither r nor r + Δr can be negative, so r stays at %v", name, v, l.Start.R)
		}
		l.SetStart(k, v)
	case EditOne, EditHeadOne:
		if sign == 0 {
			return ""
		}
		if k == KR && l.Dir().R+float64(l.OneCarry(k, sign)) < 0 {
			return fmt.Sprintf("%s 1r: r + Δr would be below 0, so 1r stays at %d", name, l.Start.One.R)
		}
		l.StepOne(k, sign)
	case EditPress:
		if k == KR && l.Start.R+float64(l.Delta.R+sign*l.Increment.R) < 0 {
			return fmt.Sprintf("%s Δr: r + Δr would be below 0, so Δr stays at %d", name, l.Delta.R)
		}
		l.Step(k, sign)
	case EditInc, EditHeadInc:
		inc := *l.Increment.At(k)
		v := inc + sign
		if m.IsTyped {
			if m.Typed != math.Trunc(m.Typed) {
				return fmt.Sprintf("%s nΔ %v: not a whole number, so it stays at %d", name, m.Typed, inc)
			}
			v = int(m.Typed)
		}
		if k == KR && v <= 0 {
			return fmt.Sprintf("%s nΔr %d: it must stay above 0, so it stays at %d", name, v, inc)
		}
		*l.Increment.At(k) = v
	case EditHead:
		return headLink(l, m, name)
	}
	return ""
}

func headLink(l *Link, m StepEdit, name string) string {
	refused := ""
	stepR := func(sign int) {
		switch {
		case m.FixedR:
			refused = fmt.Sprintf("%s Δr: a node it joins has more than one incoming link, so no radius it is part of can change; Δr stays at %v", name, l.Delta.R)
		case l.Start.R+float64(l.Delta.R+sign*l.Increment.R) < 0:
			refused = fmt.Sprintf("%s Δr: r + Δr would be below 0, so Δr stays at %v", name, l.Delta.R)
		default:
			l.Step(KR, sign)
		}
	}
	switch m.K {
	case KBoth:
		l.Step(KPhi, m.Sign*way(l, m.Locks, KPhi))
		l.Step(KTheta, m.Sign*way(l, m.Locks, KTheta))
		stepR(m.Sign * way(l, m.Locks, KR))
	case KR:
		stepR(m.Sign)
	default:
		l.Step(m.K, m.Sign)
	}
	return refused
}
