package Builder

import (
	"path/filepath"
	"slices"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func stepLabelAt(dir string, n int) string {
	var name string
	if Links.ReadIfExists(filepath.Join(stepDir(dir, n), FileStepName), &name) && name != "" {
		return "step " + strconv.Itoa(n) + " · " + name
	}
	return "step " + strconv.Itoa(n)
}

func (m Model) StepShown() int {
	ns := m.stepNums()
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
	ns := m.stepNums()
	i := slices.Index(ns, was)
	m.ShownStep = ns[min(max(i+sign, 0), len(ns)-1)]
}

func (m *Model) Touch(a Act) {
	if a.Node != "" && a.To != "" && (a.I > 0 || a.I == startPart) {
		m.Touched = StepEdge{Step: a.I, From: a.Node, To: a.To}
	}
}
