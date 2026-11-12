package Builder

import (
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const OpStepPick = "step-pick"

type StepEdge struct {
	Step     int
	From, To string
}

func (m Model) StepPicked(step int, l Links.Link) bool {
	c := m.StepLinks[StepEdge{step, l.From, l.To}]
	return c != nil && c.Picked
}

func (m *Model) ToggleStepPick(step int, from, to string) string {
	l := m.StepLinks[StepEdge{step, from, to}]
	if l == nil {
		return fmt.Sprintf("step %d %s→%s: this step has no copy of the link yet, so it can't be picked", step, from, to)
	}
	if !l.Picked {
		for o, c := range m.StepLinks {
			if c.Picked && o.Step == step && o.To == to && o.From != from {
				return fmt.Sprintf("step %d %s→%s: %s already has a picked link into it in step %d, from %s", step, from, to, to, step, o.From)
			}
		}
	}
	l.Picked = !l.Picked
	return ""
}

func (m Model) StepPickChanges(step int) (unpick, pick []Links.Link) {
	for _, l := range m.allLinks() {
		want := m.StepPicked(step, l)
		switch {
		case l.Picked && !want:
			unpick = append(unpick, l)
		case !l.Picked && want:
			pick = append(pick, l)
		}
	}
	return unpick, pick
}
