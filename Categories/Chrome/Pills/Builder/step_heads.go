package Builder

import (
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
			s = ReadStep(stepDir(m.lvl(), step))
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

type HeadEdit struct {
	To   string
	Edit Links.StepEdit
}

func (m *Model) stepHeadEdits(step int, op, k string, sign int, locks Links.Lock) []HeadEdit {
	if sign == 0 {
		return nil
	}
	var out []HeadEdit
	many := m.ManyIn()
	for _, e := range m.stepActive(step) {
		fixed := slices.Contains(many, e.To) || slices.Contains(many, e.From)
		out = append(out, HeadEdit{To: e.To, Edit: Links.StepEdit{Step: step, From: e.From, Op: op, K: k, Sign: sign, Locks: locks, FixedR: fixed}})
	}
	return out
}

func (m *Model) StepHeadEdit(a Act, sign int) []HeadEdit {
	st := m.StepHead(a.I)
	switch a.Op {
	case OpAllLocks:
		st.ToggleBoth()
		return nil
	case OpAllInc:
		return m.stepHeadEdits(a.I, Links.EditHeadInc, a.K, sign, Links.Lock{})
	case OpAllOne:
		return m.stepHeadEdits(a.I, Links.EditHeadOne, a.K, sign, Links.Lock{})
	case OpAllLock:
		way, on := st.At(a.K)
		switch {
		case sign == 0:
			*on = !*on
			st.Both = st.Both && !*on
		case *on:
			*way = sign
			return m.stepHeadEdits(a.I, Links.EditHead, a.K, sign, st.Locks())
		default:
			return m.stepHeadEdits(a.I, Links.EditHead, a.K, sign, st.Locks())
		}
		if *way == 0 {
			*way = 1
		}
		return nil
	}
	if sign == 0 {
		sign = a.Sign
	}
	return m.stepHeadEdits(a.I, Links.EditHead, Links.KBoth, sign, st.Locks())
}

func (m Model) writeStepHeads() error {
	for step, s := range m.StepHeads {
		if err := WriteStep(stepDir(m.lvl(), step), *s); err != nil {
			return err
		}
	}
	return nil
}
