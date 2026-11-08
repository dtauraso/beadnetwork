package Builder

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const DirStepHeads = "heads"

type StepNode struct {
	Step int
	ID   string
}

func stepHeadDir(stepDir string, step int, id string) string {
	return filepath.Join(stepDir, DirStepHeads, "step-"+strconv.Itoa(step), id)
}

func (m *Model) StepHead(step int, id string) *Step {
	if m.StepHeads == nil {
		m.StepHeads = map[StepNode]*Step{}
	}
	k := StepNode{step, id}
	if m.StepHeads[k] == nil {
		s := Step{}
		if m.StepDir != "" {
			s = ReadStep(stepHeadDir(m.StepDir, step, id))
		}
		m.StepHeads[k] = &s
	}
	return m.StepHeads[k]
}

func (m Model) headOf(step int, id string) Step {
	if step == 0 {
		return m.StepOf(id)
	}
	if s := m.StepHeads[StepNode{step, id}]; s != nil {
		return *s
	}
	return Step{}
}

func IsStepHead(a Act) bool {
	return a.I > 0 && (a.Op == OpAllLocks || a.Op == OpAllBoth || a.Op == OpAllLock || a.Op == OpAllInc)
}

func (m *Model) stepHeadIncrements(step int, from, k string, sign int) (string, []StepEdge) {
	if sign == 0 {
		return "", nil
	}
	per := float64(m.ScaleOf(from))
	if k == Links.KR {
		per = 1
	}
	refused := ""
	var moved []StepEdge
	for _, n := range m.linksFrom(from) {
		e := StepEdge{step, n.From, n.To}
		l := m.StepLinks[e]
		if l == nil {
			continue
		}
		inc := l.Increment.At(k)
		v := *inc*per + float64(sign)
		if k == Links.KR && v <= 0 {
			refused = fmt.Sprintf("step %d %s→%s nΔr: it would be %v, and it must stay above 0, so it stays at %v", step, n.From, n.To, v, *inc)
			continue
		}
		*inc = v / per
		moved = append(moved, e)
	}
	return refused, moved
}

func (m *Model) StepHeadEdit(a Act, sign int) (string, []StepEdge) {
	st := m.StepHead(a.I, a.Node)
	k := a.K
	switch a.Op {
	case OpAllLocks:
		st.ToggleBoth()
		return "", nil
	case OpAllInc:
		return m.stepHeadIncrements(a.I, a.Node, k, sign)
	case OpAllLock:
		way, on := st.At(a.K)
		switch {
		case sign == 0:
			*on = !*on
			st.Both = st.Both && !*on
		case *on:
			*way = sign
			return m.stepHeadLinks(a.I, a.Node, k, sign, st.Locks())
		default:
			return m.stepHeadLinks(a.I, a.Node, k, sign, st.Locks())
		}
		if *way == 0 {
			*way = 1
		}
		return "", nil
	}
	if sign == 0 {
		sign = a.Sign
	}
	return m.stepHeadLinks(a.I, a.Node, Links.KBoth, sign, st.Locks())
}

func (m *Model) stepHeadLinks(step int, from, k string, sign int, locks Links.Lock) (string, []StepEdge) {
	refused := ""
	var moved []StepEdge
	if sign == 0 {
		return "", nil
	}
	for _, n := range m.linksFrom(from) {
		e := StepEdge{step, n.From, n.To}
		l := m.StepLinks[e]
		if l == nil {
			continue
		}
		stepR := func(sign int) {
			name := fmt.Sprintf("step %d %s→%s Δr", step, n.From, n.To)
			switch {
			case slices.Contains(m.ManyIn(), n.To) || slices.Contains(m.ManyIn(), n.From):
				refused = fmt.Sprintf("%s: a node it joins has more than one incoming link, so no radius it is part of can change; Δr stays at %v", name, l.Delta.R)
			case l.Start.R+l.Delta.R+float64(sign)*l.Increment.R < 0:
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
	at := func(x Links.Lock) int { return map[string]int{Links.KPhi: x.Phi, Links.KTheta: x.Theta, Links.KR: x.R}[k] }
	if v := at(l.Lock); v != 0 {
		return v
	}
	if v := at(all); v != 0 {
		return v
	}
	return 1
}

func (m Model) writeStepHeads() error {
	for k, s := range m.StepHeads {
		dir := stepHeadDir(m.StepDir, k.Step, k.ID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := WriteStep(dir, *s); err != nil {
			return err
		}
	}
	return nil
}
