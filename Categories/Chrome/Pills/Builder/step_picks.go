package Builder

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
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
		if err := Sheet.WriteScalar(filepath.Join(stepDir(m.lvl(), step), FileStepOpen), v); err != nil {
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
		sl, ok := m.gridLink(e)
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
	for _, n := range m.stepNums() {
		if v := Sheet.LoadScalar(filepath.Join(stepDir(m.lvl(), n), FileStepOpen), -1, 0); v >= 0 {
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
	if l, ok := m.gridLink(StepEdge{a.I, a.Node, a.To}); !ok || l.Picked {
		return ""
	}
	path := []string{a.To, a.Node}
	for cur := a.Node; len(path) <= len(m.Nodes)+1; {
		next := ""
		for _, l := range m.linksTo(cur) {
			if sl, ok := m.gridLink(StepEdge{a.I, l.From, cur}); ok && sl.Picked {
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
	c, ok := m.gridLink(StepEdge{step, l.From, l.To})
	return ok && c.Picked
}
