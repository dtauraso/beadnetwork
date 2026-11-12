package Builder

import (
	"path/filepath"
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
	m.StepClosed = false
	return m.rereadSteps()
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
	m.SubSteps = map[int]int{}
	for _, n := range m.stepNums() {
		m.SubSteps[n] = len(ReadStepNumbers(stepDir(m.lvl(), n)))
	}
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

func subStepsCell(m Model, n int) Cell {
	tip := "open this step's own steps"
	if k := m.SubSteps[n]; k > 0 {
		return Button("→ "+strconv.Itoa(k)+map[bool]string{true: " step", false: " steps"}[k == 1], Act{Op: OpStepEnter, I: n}, false).Tipped(tip)
	}
	return Button("→", Act{Op: OpStepEnter, I: n}, false).Tipped(tip + "; it has none yet, so going in makes its step 1")
}

func trailCells(m Model) []Cell {
	name := "sequence"
	for _, q := range m.Seqs {
		if q.K == m.OpenSeq {
			name = q.Name
		}
	}
	cells := []Cell{Button(name, Act{Op: OpLevelGo, I: stepIn(m.Trail, 0)}, m.Level == "").Tipped("go to the sequence's own steps")}
	for i, t := range m.Trail {
		at := filepath.Join(t.Level, DirSteps, strconv.Itoa(t.Step))
		cells = append(cells, Text("›").Dimmed(true),
			Button(t.Label, Act{Op: OpLevelGo, Node: at, I: stepIn(m.Trail, i+1)}, at == m.Level).Tipped("go to this step's own steps"))
	}
	return cells
}

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
			Button(t.Label, Act{Op: OpLevelGo, Node: t.Level, I: t.Step}, here).Tipped("open this step in the Links tab"),
		})
	}
	return Group{Rows: rows, Slides: true, Table: "step tree"}
}
