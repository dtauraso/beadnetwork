package Builder

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	OpRename  = "rename"
	OpLevelUp = "level-up"
	OpLinkRow = "link-row"

	renameStep = "step"
	renameSeq  = "seq"
)

const (
	leaf = iota
	closed
	open
)

type linkRow struct {
	id     Act
	depth  int
	open   int
	fold   Act
	in     Act
	rename Act
	name   Cell
	detail []Cell
	acts   []Cell
	more   []Cell
	table  string
}

type tree struct {
	sel    Act
	groups []Group
	rows   []linkRow
}

func itoa(n int) string { return strconv.Itoa(n) }

func chevron(r linkRow) Cell {
	switch r.open {
	case closed:
		return Text("›").Fitting("⌄").Acting(r.fold).Tipped("unfold")
	case open:
		return Text("⌄").Fitting("⌄").Acting(r.fold).Tipped("fold")
	}
	return Text("").Fitting("⌄")
}

func (t *tree) add(r linkRow) {
	t.rows = append(t.rows, r)
	cells := append([]Cell{chevron(r), r.name}, r.detail...)
	cells = append(cells, r.more...)
	g := Group{Depth: r.depth, Row: r.id, Slides: r.table != "", Table: r.table}
	if r.id == t.sel {
		cells = append(cells, r.acts...)
		g.Band, g.Trail = Sheet.BandSelected, len(r.acts)
	}
	g.Rows = [][]Cell{cells}
	t.groups = append(t.groups, g)
}

func (t *tree) put(g ...Group) { t.groups = append(t.groups, g...) }

func (m Model) linkRows() []linkRow { return linksTree(m).rows }

func (m *Model) Select(id Act) { m.Sel = id }

func (m Model) rowOf(id Act) (linkRow, bool) {
	for _, r := range m.linkRows() {
		if r.id == id {
			return r, true
		}
	}
	return linkRow{}, false
}

func (m Model) RowIn(id Act) Act {
	r, ok := m.rowOf(id)
	if !ok {
		return Act{}
	}
	if r.in.Op != "" {
		return r.in
	}
	return r.fold
}

const doubleClickMs = 400

func (m *Model) Press(id Act, now int64) Act {
	double := m.ClickRow == id && now-m.ClickAt < doubleClickMs
	m.Sel, m.ClickRow, m.ClickAt = id, id, now
	if double {
		m.ClickRow = Act{}
		return m.RowIn(id)
	}
	return Act{}
}

func (m Model) Rename(a Act, e *Sheet.Edit) {
	if c, ok := m.RenameCell(a); ok {
		Sheet.StartDraft(e, c)
	}
}

func (m *Model) LevelUp() error {
	if n := len(m.Trail); n > 0 {
		t := m.Trail[n-1]
		return m.LevelOp(Act{Op: OpLevelGo, Node: t.Level, I: t.Step})
	}
	if !m.AtSeqs && len(m.Seqs) > 0 {
		m.ToSeqs()
	}
	return nil
}

func (m Model) RenameCell(a Act) (Cell, bool) {
	switch a.K {
	case renameStep:
		return Field("", m.StepNames[a.I], Act{Op: OpStepName, I: a.I}, true, false).Wording(), true
	case renameSeq:
		return Field("", m.SeqNamed(a.I), Act{Op: OpSeqName, I: a.I}, true, false).Wording(), true
	}
	return Cell{}, false
}

func IsRowName(a Act) bool { return a.Op == OpStepName || a.Op == OpSeqName }
