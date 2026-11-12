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

const OpCardFold = "card-fold"

func (m Model) CardOpen(e StepEdge, picked bool) bool {
	if on, ok := m.CardsOpen[e]; ok {
		return on
	}
	return picked
}

func (m *Model) ToggleCard(a Act) {
	e := StepEdge{a.I, a.Node, a.To}
	l, _ := m.Link(a.Node, a.To)
	picked := l.Picked
	if a.I > 0 {
		picked = m.StepPicked(a.I, Links.Link{From: a.Node, To: a.To})
	}
	if m.CardsOpen == nil {
		m.CardsOpen = map[StepEdge]bool{}
	}
	m.CardsOpen[e] = !m.CardOpen(e, picked)
}

func (m Model) cardRow(r linkRow, e StepEdge, picked bool) linkRow {
	r.fold = Act{Op: OpCardFold, Node: e.From, To: e.To, I: e.Step}
	r.open = openOf(m.CardOpen(e, picked))
	return r
}

func (m Model) StepPicked(step int, l Links.Link) bool {
	c, ok := m.gridLink(StepEdge{step, l.From, l.To})
	return ok && c.Picked
}
