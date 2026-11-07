package Builder

const OpStepFold = "step-fold"

func (m Model) StepFolded(step int, id string) bool {
	if f, ok := m.StepFolds[StepNode{step, id}]; ok {
		return f
	}
	for _, l := range m.stepLinksFrom(step, id) {
		if m.StepPicked(step, l) {
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

func foldTitle(folded bool, title string) string {
	if folded {
		return "▶ " + title
	}
	return "▼ " + title
}
