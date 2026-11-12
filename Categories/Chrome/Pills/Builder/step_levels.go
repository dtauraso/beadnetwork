package Builder

import (
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpStepEnter = "step-enter"
	OpLevelGo   = "level-go"
)

type TreeRow struct {
	Level string
	Step  int
	Depth int
	Label string
}

type levelStep struct {
	Level string
	Step  int
}

func IsLevelOp(a Act) bool { return a.Op == OpStepEnter || a.Op == OpLevelGo }

func (m *Model) LevelOp(a Act) error {
	if m.Root == "" {
		return nil
	}
	if err := m.WriteSteps(); err != nil {
		return err
	}
	if a.Op == OpStepEnter {
		m.Level, m.ShownStep = filepath.Join(m.Level, DirSteps, strconv.Itoa(a.I)), 1
	} else {
		m.Level = a.Node
		if a.I > 0 {
			m.ShownStep = a.I
		}
	}
	m.StepClosed, m.AtSeqs = false, false
	err := m.rereadSteps()
	m.Sel = Act{Op: OpStepShow, I: m.StepShown()}
	return err
}

func (m Model) ancestors() []levelStep {
	var out []levelStep
	parts := strings.Split(filepath.ToSlash(m.Level), "/")
	for i := 1; i < len(parts); i += 2 {
		n, err := strconv.Atoi(parts[i])
		if err != nil || parts[i-1] != DirSteps {
			continue
		}
		out = append(out, levelStep{Level: filepath.Join(parts[:i-1]...), Step: n})
	}
	return out
}

func stepLabelAt(dir string, n int) string {
	var name string
	if Links.ReadIfExists(filepath.Join(stepDir(dir, n), FileStepName), &name) && name != "" {
		return "step " + strconv.Itoa(n) + " · " + name
	}
	return "step " + strconv.Itoa(n)
}

func (m *Model) readLevels() {
	m.Trail = nil
	for _, a := range m.ancestors() {
		m.Trail = append(m.Trail, TreeRow{Level: a.Level, Step: a.Step, Label: stepLabelAt(filepath.Join(m.Root, a.Level), a.Step)})
	}
	m.Tree = nil
	m.readTree("", 0)
}

func (m *Model) readTree(level string, depth int) {
	dir := filepath.Join(m.Root, level)
	for _, n := range ReadStepNumbers(dir) {
		m.Tree = append(m.Tree, TreeRow{Level: level, Step: n, Depth: depth, Label: stepLabelAt(dir, n)})
		m.readTree(filepath.Join(level, DirSteps, strconv.Itoa(n)), depth+1)
	}
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
	if sign == 0 {
		if a.I == m.StepShown() && !m.StepClosed {
			m.StepClosed = true
			return
		}
		m.ShownStep, m.StepClosed = a.I, false
		m.ShownStep = m.StepShown()
		return
	}
	ns := m.stepNums()
	i := slices.Index(ns, m.StepShown())
	m.ShownStep = ns[min(max(i+sign, 0), len(ns)-1)]
}

func (m *Model) Touch(a Act) {
	if a.Node != "" && a.To != "" && a.I > 0 {
		m.Touched = StepEdge{Step: a.I, From: a.Node, To: a.To}
	}
}

func (m Model) stepOpen(n int) bool { return n == m.StepShown() && !m.StepClosed }

func (m Model) TopStep() int { return stepIn(m.Trail, 0) }

func stepIn(trail []TreeRow, i int) int {
	if i < len(trail) {
		return trail[i].Step
	}
	return 0
}

func stepTree(m Model) Group {
	rows := [][]Cell{{Text("steps").Dimmed(true)}}
	if len(m.Tree) == 0 {
		rows = append(rows, []Cell{Text("this sequence has no step folders yet")})
	}
	for _, t := range m.Tree {
		here := t.Level == m.Level && t.Step == m.StepShown()
		rows = append(rows, []Cell{
			Text(strings.Repeat("    ", t.Depth)),
			LitText(t.Label, Act{Op: OpLevelGo, Node: t.Level, I: t.Step}, here).Tipped("open this step in the Links tab"),
		})
	}
	return Group{Rows: rows, Slides: true, Table: "step tree"}
}
