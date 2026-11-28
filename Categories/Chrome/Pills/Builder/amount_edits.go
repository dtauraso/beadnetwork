package Builder

import (
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func (m Model) stepLink(e StepEdge) (Links.Link, bool) {
	l, ok := m.Link(e.From, e.To)
	if !ok {
		return Links.Link{}, false
	}
	a, ok := m.AmountOf(e.Step, Edge{From: e.From, To: e.To})
	if !ok {
		return Links.Link{}, false
	}
	l.Delta, l.Increment, l.Picked = a.D, a.Inc, a.Picked
	return l, true
}

var editOps = map[string]string{
	OpLinkStart: Links.EditStart, OpLinkOne: Links.EditOne,
}

func IsStartEdit(a Act) bool { return a.I == startPart && (a.Op == OpLinkStart || a.Op == OpLinkOne) }

func IsAmountEdit(a Act) bool {
	return a.I > 0 && (a.Op == OpLinkPress || a.Op == OpIncrement || a.Op == OpStepPick)
}

const (
	startPart   = -1
	OpStartFold = "start-fold"
)

type HeadEdit = Links.Envelope

func (m Model) refuseStep(step int) string {
	return fmt.Sprintf("step %d holds inner steps, so it holds no amounts of its own; open an inner step to change them", step)
}
func (m Model) name(step int, e Edge) string {
	return fmt.Sprintf("step %d %s→%s", step, e.From, e.To)
}

func (m *Model) AmountEdit(a Act, sign int, typed *float64) (string, error) {
	step, e := a.I, Edge{From: a.Node, To: a.To}
	cur, ok := m.AmountOf(step, e)
	if !ok {
		return m.refuseStep(step), nil
	}
	p := m.stepPath(step)
	name := m.name(step, e)
	switch a.Op {
	case OpStepPick:
		cur.Picked = !cur.Picked
		if cur.Picked {
			if err := m.movePick(step, e); err != nil {
				return "", err
			}
		}
	case OpLinkPress:
		if sign == 0 {
			sign = a.Sign
		}
		if !cur.Picked {
			return fmt.Sprintf("%s Δ: %s→%s is not picked in this step, so its Δ does not move %s; Δ stays at %+v", name, e.From, e.To, e.To, cur.D), nil
		}
		cur.Step(a.K, sign)
		if a.K == Links.KR {
			l, _ := m.Link(e.From, e.To)
			if end := l.Start.R + float64(m.Through(step, e).R-m.amountAt(p, e).D.R+cur.D.R); end < 0 {
				cur.D.R += int(math.Ceil(-end - 1e-9))
			}
		}
	case OpIncrement:
		if refused := incEdit(&cur, a.K, sign, typed, name); refused != "" {
			return refused, nil
		}
	}
	m.ShownStep, m.StepClosed = step, false
	return "", m.setAmount(p, e, cur)
}

func (m *Model) movePick(step int, e Edge) error {
	p := m.stepPath(step)
	for _, l := range m.linksTo(e.To) {
		o := Edge{From: l.From, To: l.To}
		if o == e {
			continue
		}
		if a := m.amountAt(p, o); a.Picked {
			a.Picked = false
			m.Refused = fmt.Sprintf("moved the pick from %s→%s to %s→%s: one picked link into a node at a time", o.From, o.To, e.From, e.To)
			if err := m.setAmount(p, o, a); err != nil {
				return err
			}
		}
	}
	return nil
}

func incEdit(a *Amount, k string, sign int, typed *float64, name string) string {
	inc := *a.Inc.At(k)
	v := inc + sign
	if typed != nil {
		if *typed != math.Trunc(*typed) {
			return fmt.Sprintf("%s nΔ %v: not a whole number, so it stays at %d", name, *typed, inc)
		}
		v = int(*typed)
	}
	if k == Links.KR && v <= 0 {
		return fmt.Sprintf("%s nΔr %d: it must stay above 0, so it stays at %d", name, v, inc)
	}
	*a.Inc.At(k) = v
	return ""
}

func way(l Links.Link, all Links.Lock, k string) int {
	if v := *l.Lock.At(k); v != 0 {
		return v
	}
	if v := *all.At(k); v != 0 {
		return v
	}
	return 1
}

func (m *Model) headEdit(step int, op, k string, sign int, locks Links.Lock) (string, []Links.Envelope, error) {
	if sign == 0 {
		return "", nil, nil
	}
	if m.StepIsParent(step) {
		return m.refuseStep(step), nil, nil
	}
	p := m.stepPath(step)
	refused := ""
	var out []Links.Envelope
	for _, e := range m.stepActive(step) {
		l, _ := m.Link(e.From, e.To)
		edge := Edge{From: e.From, To: e.To}
		a := m.amountAt(p, edge)
		name := m.name(step, edge)
		stepR := func(sign int) {
			switch {
			case m.FixedR(l):
				refused = fmt.Sprintf("%s Δr: a node it joins has more than one incoming link, so no radius it is part of can change; Δr stays at %v", name, a.D.R)
			case l.Start.R+float64(m.Through(step, edge).R+sign*a.Inc.R) < 0:
				refused = fmt.Sprintf("%s Δr: r + Δr would be below 0, so Δr stays at %v", name, a.D.R)
			default:
				a.Step(Links.KR, sign)
			}
		}
		switch op {
		case OpAllOne:
			out = append(out, Links.Envelope{To: e.To, Msg: Links.StartEdit{From: e.From, Op: Links.EditOne, K: k, Sign: sign}})
			continue
		case OpAllInc:
			if r := incEdit(&a, k, sign, nil, name); r != "" {
				refused = r
			}
		default:
			switch k {
			case Links.KBoth:
				a.Step(Links.KPhi, sign*way(l, locks, Links.KPhi))
				a.Step(Links.KTheta, sign*way(l, locks, Links.KTheta))
				stepR(sign * way(l, locks, Links.KR))
			case Links.KR:
				stepR(sign)
			default:
				a.Step(k, sign)
			}
		}
		if err := m.setAmount(p, edge, a); err != nil {
			return refused, out, err
		}
	}
	m.ShownStep, m.StepClosed = step, false
	return refused, out, nil
}

func (m *Model) StartEdit(a Act, sign int, typed *float64) Links.Envelope {
	e := Links.StartEdit{From: a.Node, Op: editOps[a.Op], K: a.K, Sign: sign}
	if typed != nil {
		e.Typed, e.IsTyped = *typed, true
	}
	return Links.Envelope{To: a.To, Msg: e}
}
