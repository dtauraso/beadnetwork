package Builder

import (
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
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
	return a.I > 0 && (a.Op == OpAllLocks || a.Op == OpAllBoth || a.Op == OpAllLock || a.Op == OpAllLockCycle || a.Op == OpAllInc || a.Op == OpAllOne || a.Op == OpAllScale || a.Op == OpAllPlay)
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
	case OpAllLockCycle:
		way, on := st.At(a.K)
		switch {
		case !*on:
			*on, *way = true, 1
		case *way > 0:
			*way = -1
		default:
			*on, *way = false, 1
		}
		st.Both = st.Both && !*on
		return nil
	case OpAllLocks:
		st.ToggleBoth()
		return nil
	case OpAllScale:
		st.Scale = max(1, st.PlayScale()+sign)
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

const (
	OpAllPlay  = "all-play"
	OpAllScale = "all-scale"
)

type Play struct {
	Step int
	K, N int
	At   int64
}

func (m Model) PlayPresses(step int) int {
	presses := 0
	for _, e := range m.stepActive(step) {
		if l, ok := m.gridLink(e); ok {
			presses = max(presses, l.Presses())
		}
	}
	return presses
}

func (m Model) playEdits(step, press, of int) []HeadEdit {
	var out []HeadEdit
	for _, e := range m.stepActive(step) {
		out = append(out, HeadEdit{To: e.To, Edit: Links.StepEdit{Step: step, From: e.From, Op: Links.EditPlay, Press: press, Of: of}})
	}
	return out
}

func (m Model) Playing(step int) bool { return m.Play != nil && m.Play.Step == step }

func playOf(m Model, step int) LinksBoard.Lock {
	glyph := "▶"
	if m.Playing(step) {
		glyph = "■ " + itoa(m.Play.K) + "/" + itoa(m.Play.N)
	}
	return LinksBoard.Lock{Label: glyph, Axis: "×" + itoa(m.headOf(step).PlayScale()), Value: itoa(m.headOf(step).PlayScale()), On: m.Playing(step),
		Cycle: Act{Op: OpAllPlay, I: step}, Press: Act{Op: OpAllScale, I: step}}
}

func (m *Model) StartPlay(step int, now int64) {
	m.Play = &Play{Step: step, N: max(1, m.PlayPresses(step)*m.headOf(step).PlayScale()), At: now + 2}
}

func (m *Model) StopPlay() (end []HeadEdit, was bool) {
	p := m.Play
	if p == nil {
		return nil, false
	}
	m.Play = nil
	if p.K > 0 {
		end = m.playEdits(p.Step, p.N, p.N)
	}
	return end, true
}

func (m *Model) PlayTick(now int64) ([]HeadEdit, bool) {
	p := m.Play
	if p == nil || now < p.At || m.Walking() {
		return nil, false
	}
	p.At = now + 1
	edits := m.playEdits(p.Step, p.K, p.N)
	if p.K >= p.N {
		m.Play = nil
	}
	p.K++
	return edits, true
}
