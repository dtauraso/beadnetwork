package Builder

import (
	"fmt"
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type StepNode struct {
	Step int
	ID   string
}

func (m *Model) StepHead(step int) *Step {
	if m.StepHeads == nil {
		m.StepHeads = map[int]*Step{}
	}
	if m.StepHeads[step] == nil {
		s := Step{}
		if m.Root != "" {
			s = ReadStep(stepDir(m.Root, step))
		}
		m.StepHeads[step] = &s
	}
	return m.StepHeads[step]
}

func (m Model) headOf(step int) Step {
	if s := m.StepHeads[step]; s != nil {
		return *s
	}
	return Step{}
}

func (m Model) stepActive(step int) []StepEdge {
	var out []StepEdge
	for _, l := range m.allLinks() {
		e := StepEdge{step, l.From, l.To}
		if m.StepPicked(step, l) {
			out = append(out, e)
		}
	}
	return out
}

func IsStepHead(a Act) bool {
	return a.I > 0 && (a.Op == OpAllLocks || a.Op == OpAllBoth || a.Op == OpAllLock || a.Op == OpAllInc || a.Op == OpAllOne)
}

func (m *Model) stepHeadIncrements(step int, k string, sign int) (string, []StepEdge) {
	if sign == 0 {
		return "", nil
	}
	refused := ""
	var moved []StepEdge
	for _, e := range m.stepActive(step) {
		l := m.StepLinks[e]
		inc := l.Increment.At(k)
		v := *inc + sign
		if k == Links.KR && v <= 0 {
			refused = fmt.Sprintf("step %d %s→%s nΔr: it would be %d, and it must stay above 0, so it stays at %d", step, e.From, e.To, v, *inc)
			continue
		}
		*inc = v
		moved = append(moved, e)
	}
	return refused, moved
}

func (m *Model) stepHeadOnes(step int, k string, sign int) (string, []StepEdge) {
	if sign == 0 {
		return "", nil
	}
	refused := ""
	var moved []StepEdge
	for _, e := range m.stepActive(step) {
		l := m.StepLinks[e]
		if k == Links.KR && l.Dir().R+float64(l.OneCarry(k, sign)) < 0 {
			refused = fmt.Sprintf("step %d %s→%s 1r: r + Δr would be below 0, so 1r stays at %d", step, e.From, e.To, l.Start.One.R)
			continue
		}
		l.StepOne(k, sign)
		moved = append(moved, e)
	}
	return refused, moved
}

func (m *Model) StepHeadEdit(a Act, sign int) (string, []StepEdge) {
	st := m.StepHead(a.I)
	k := a.K
	switch a.Op {
	case OpAllLocks:
		st.ToggleBoth()
		return "", nil
	case OpAllInc:
		return m.stepHeadIncrements(a.I, k, sign)
	case OpAllOne:
		return m.stepHeadOnes(a.I, k, sign)
	case OpAllLock:
		way, on := st.At(a.K)
		switch {
		case sign == 0:
			*on = !*on
			st.Both = st.Both && !*on
		case *on:
			*way = sign
			return m.stepHeadLinks(a.I, k, sign, st.Locks())
		default:
			return m.stepHeadLinks(a.I, k, sign, st.Locks())
		}
		if *way == 0 {
			*way = 1
		}
		return "", nil
	}
	if sign == 0 {
		sign = a.Sign
	}
	return m.stepHeadLinks(a.I, Links.KBoth, sign, st.Locks())
}

func (m *Model) stepHeadLinks(step int, k string, sign int, locks Links.Lock) (string, []StepEdge) {
	refused := ""
	var moved []StepEdge
	if sign == 0 {
		return "", nil
	}
	for _, e := range m.stepActive(step) {
		l := m.StepLinks[e]
		stepR := func(sign int) {
			name := fmt.Sprintf("step %d %s→%s Δr", step, e.From, e.To)
			switch {
			case slices.Contains(m.ManyIn(), e.To) || slices.Contains(m.ManyIn(), e.From):
				refused = fmt.Sprintf("%s: a node it joins has more than one incoming link, so no radius it is part of can change; Δr stays at %v", name, l.Delta.R)
			case l.Start.R+float64(l.Delta.R+sign*l.Increment.R) < 0:
				refused = fmt.Sprintf("%s: r + Δr would be below 0, so Δr stays at %v", name, l.Delta.R)
			default:
				l.Step(Links.KR, sign)
			}
		}
		switch k {
		case Links.KBoth:
			l.Step(Links.KPhi, sign*wayOf(l, locks, Links.KPhi))
			l.Step(Links.KTheta, sign*wayOf(l, locks, Links.KTheta))
			stepR(sign * wayOf(l, locks, Links.KR))
		case Links.KR:
			stepR(sign)
		default:
			l.Step(k, sign)
		}
		moved = append(moved, e)
	}
	return refused, moved
}

func wayOf(l *Links.Link, all Links.Lock, k string) int {
	at := func(x Links.Lock) int {
		return map[string]int{Links.KPhi: x.Phi, Links.KTheta: x.Theta, Links.KR: x.R}[k]
	}
	if v := at(l.Lock); v != 0 {
		return v
	}
	if v := at(all); v != 0 {
		return v
	}
	return 1
}

func (m Model) writeStepHeads() error {
	for step, s := range m.StepHeads {
		if err := WriteStep(stepDir(m.Root, step), *s); err != nil {
			return err
		}
	}
	return nil
}
