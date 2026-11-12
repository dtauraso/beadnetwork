package Builder

import "github.com/dtauraso/beadnetwork/Categories/Node/Links"

const (
	OpStepPick = "step-pick"
	OpStepFold = "step-fold"
)

func (m Model) StepFolded(step int, id string) bool {
	if f, ok := m.StepFolds[StepNode{step, id}]; ok {
		return f
	}
	for _, l := range m.gridLinks(step, id) {
		if step == 0 && l.Picked || m.StepPicked(step, l) {
			return false
		}
	}
	return true
}

func (m *Model) ToggleStepFold(step int, id string) {
	if m.StepFolds == nil {
		m.StepFolds = map[StepNode]bool{}
	}
	m.StepFolds[StepNode{step, id}] = !m.StepFolded(step, id)
}

type StepEdge struct {
	Step     int
	From, To string
}

func (m Model) StepPicked(step int, l Links.Link) bool {
	c, ok := m.gridLink(StepEdge{step, l.From, l.To})
	return ok && c.Picked
}
