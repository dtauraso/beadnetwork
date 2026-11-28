package Builder

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet/Scalar"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/StepStore"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const OpIncrement = "increment"

func (m Model) Picked() []Links.Link {
	var out []Links.Link
	for _, l := range m.allLinks() {
		if l.Picked {
			out = append(out, l)
		}
	}
	return out
}

const (
	OpAllBoth      = "all-both"
	OpAllLock      = "all-lock"
	OpAllLockCycle = "all-lock-cycle"
	OpAllLocks     = "all-locks"
	OpAllInc       = "all-increment"
	OpAllOne       = "all-one"
)

const (
	OpStepPick   = "step-pick"
	OpStepFold   = "step-fold"
	FileStepOpen = "open.bin"
)

func (m Model) StepOpen(step int) bool {
	if on, ok := m.StepsOpen[StepNode{step, m.Level}]; ok {
		return on
	}
	return step == m.StepShown()
}

func (m *Model) setStepOpen(step int, on bool) {
	if m.StepsOpen == nil {
		m.StepsOpen = map[StepNode]bool{}
	}
	m.StepsOpen[StepNode{step, m.Level}] = on
	if m.Root != "" && step > 0 {
		v := int32(0)
		if on {
			v = 1
		}
		if err := Scalar.Write(filepath.Join(StepStore.StepDir(m.Lvl(), step), FileStepOpen), v); err != nil {
			fmt.Fprintf(os.Stderr, "step_picks: step %d open: %v\n", step, err)
		}
	}
}

type EdgeState struct {
	From, To     string
	Open, Picked bool
}

func (m Model) EdgeStates() []EdgeState {
	step := m.StepShown()
	if m.NetworkOpen || step <= 0 {
		return nil
	}
	var out []EdgeState
	for _, l := range m.allLinks() {
		e := StepEdge{step, l.From, l.To}
		sl, ok := m.stepLink(e)
		if !ok {
			continue
		}
		out = append(out, EdgeState{From: l.From, To: l.To, Picked: sl.Picked, Open: m.StepOpen(step) && m.CardOpen(e)})
	}
	return out
}

func (m *Model) readStepsOpen() {
	if m.StepsOpen == nil {
		m.StepsOpen = map[StepNode]bool{}
	}
	for _, n := range m.Numbers() {
		if v := Scalar.Load(filepath.Join(StepStore.StepDir(m.Lvl(), n), FileStepOpen), -1, 0); v >= 0 {
			m.StepsOpen[StepNode{n, m.Level}] = v == 1
		}
	}
}

func (m *Model) ToggleStepFold(step int) { m.setStepOpen(step, !m.StepOpen(step)) }

type StepEdge struct {
	Step     int
	From, To string
}

const (
	OpCardFold     = "card-fold"
	OpLinkPickOpen = "link-pick-open"
	OpLinkGap      = "link-gap"
)

func (m *Model) PickOpen(a Act) Act {
	if !m.CardOpen(StepEdge{a.I, a.Node, a.To}) {
		m.ToggleCard(a)
	}
	a.Op = OpStepPick
	return a
}

func (m Model) PickLoop(a Act) string {
	if a.Op != OpStepPick {
		return ""
	}
	if l, ok := m.stepLink(StepEdge{a.I, a.Node, a.To}); !ok || l.Picked {
		return ""
	}
	path := []string{a.To, a.Node}
	for cur := a.Node; len(path) <= len(m.Nodes)+1; {
		next := ""
		for _, l := range m.linksTo(cur) {
			if sl, ok := m.stepLink(StepEdge{a.I, l.From, cur}); ok && sl.Picked {
				next = l.From
			}
		}
		if next == "" {
			return ""
		}
		path = append(path, next)
		if next == a.To {
			slices.Reverse(path)
			return fmt.Sprintf("step %d: picking %s→%s would close a loop of picked links (%s), where each node keeps moving the next, so it stays unpicked", a.I, a.Node, a.To, strings.Join(path, "→"))
		}
		cur = next
	}
	return ""
}

func (m Model) CardOpen(e StepEdge) bool {
	if on, ok := m.CardsOpen[e]; ok || e.Step != 0 {
		return on
	}
	l, _ := m.Link(e.From, e.To)
	return l.Picked
}

func (m *Model) ToggleCard(a Act) {
	e := StepEdge{a.I, a.Node, a.To}
	open := m.CardOpen(e)
	if m.CardsOpen == nil {
		m.CardsOpen = map[StepEdge]bool{}
	}
	m.CardsOpen[e] = !open
}

func (m Model) StepPicked(step int, l Links.Link) bool {
	c, ok := m.stepLink(StepEdge{step, l.From, l.To})
	return ok && c.Picked
}

func (m Model) StepShown() int {
	ns := m.Numbers()
	if slices.Contains(ns, m.ShownStep) {
		return m.ShownStep
	}
	if m.ShownStep > ns[len(ns)-1] {
		return ns[len(ns)-1]
	}
	return ns[0]
}

func (m *Model) ShowStep(a Act, sign int) {
	was := m.StepShown()
	if _, ok := m.StepsOpen[StepNode{was, m.Level}]; !ok && was > 0 {
		m.setStepOpen(was, m.StepOpen(was))
	}
	if sign == 0 {
		if a.I == was && !m.StepClosed {
			m.StepClosed = true
			return
		}
		m.ShownStep, m.StepClosed = a.I, false
		m.ShownStep = m.StepShown()
		if m.ShownStep > 0 {
			m.setStepOpen(m.ShownStep, true)
		}
		return
	}
	ns := m.Numbers()
	i := slices.Index(ns, was)
	m.ShownStep = ns[min(max(i+sign, 0), len(ns)-1)]
}
