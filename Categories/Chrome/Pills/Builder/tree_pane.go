package Builder

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpTreeFind    = "tree-find"
	OpTreeName    = "tree-name"
	OpTreeAdd     = "tree-add"
	OpTreeInner   = "tree-inner"
	OpTreeCopy    = "tree-copy"
	OpTreeDrop    = "tree-drop"
	OpTreeEarlier = "tree-earlier"
	OpTreeLater   = "tree-later"

	menuTree   = "tree"
	renamePath = "path"
)

func IsTreeStepOp(a Act) bool {
	switch a.Op {
	case OpTreeAdd, OpTreeInner, OpTreeCopy, OpTreeDrop, OpTreeEarlier, OpTreeLater:
		return true
	}
	return false
}

var dirOpOf = map[string]string{OpTreeAdd: OpStepAdd, OpTreeCopy: OpStepCopy, OpTreeDrop: OpStepDrop, OpTreeEarlier: OpStepEarlier, OpTreeLater: OpStepLater}

func DirOpOf(op string) string { return dirOpOf[op] }

func (m Model) PathName(p []int) string {
	var name string
	Links.ReadIfExists(filepath.Join(m.pathDir(p), FileStepName), &name)
	return name
}

func (m *Model) NamePath(p []int, name string) string {
	if len(p) == 0 || m.Scene == "" {
		return "this scene keeps no steps, so the name was not kept"
	}
	if m.pathDir(p[:len(p)-1]) == filepath.Join(m.Root, m.Level) {
		return m.NameStep(p[len(p)-1], name)
	}
	if err := Links.WriteAtomic(filepath.Join(m.pathDir(p), FileStepName), strings.TrimSpace(name)); err != nil {
		return "step " + PathKey(p) + ": " + err.Error()
	}
	m.ReadTree()
	return ""
}

func (m *Model) AddInner(p []int) ([]int, error) {
	if err := m.WriteSteps(); err != nil {
		return nil, err
	}
	dir := m.pathDir(p)
	n := 1
	if ns := ReadStepNumbers(dir); len(ns) > 0 {
		n = ns[len(ns)-1] + 1
	}
	if err := Links.WriteAtomic(filepath.Join(dir, FileTreeOpen), true); err != nil {
		return nil, err
	}
	return append(slices.Clone(p), n), os.MkdirAll(stepDir(dir, n), 0o755)
}

func (m Model) treeMenu(r StepRow) ([]Board.Chip, Board.Chip) {
	key := PathKey(r.Path)
	items := []Board.Chip{
		{Label: "Open", Act: Act{Op: OpTreeGo, Node: key}},
		{Label: "Rename", Act: Act{Op: OpRename, K: renamePath, Node: key}},
		{Label: "Add step after", Act: Act{Op: OpTreeAdd, Node: key}},
		{Label: "Add inner step", Act: Act{Op: OpTreeInner, Node: key}, Note: "inside this one"},
		{Label: "Duplicate", Act: Act{Op: OpTreeCopy, Node: key}},
		{Label: "Move earlier", Act: Act{Op: OpTreeEarlier, Node: key}},
		{Label: "Move later", Act: Act{Op: OpTreeLater, Node: key}},
	}
	return items, Board.Chip{Label: "Delete step", Act: Act{Op: OpTreeDrop, Node: key}}
}

func (m Model) treeRows() []LinksBoard.TreeRow {
	var out []LinksBoard.TreeRow
	for _, r := range m.Tree {
		key := PathKey(r.Path)
		t := LinksBoard.TreeRow{Label: r.Label, Depth: r.Depth, Kids: r.Kids, Open: r.Open, Here: r.Here, Shown: r.Shown,
			Go: Act{Op: OpTreeGo, Node: key}, Fold: Act{Op: OpTreeFold, Node: key}, Menu: Act{Op: OpMenu, K: menuTree, Node: key},
			Name: Act{Op: OpTreeName, Node: key}}
		if t.MenuOpen = m.Menu == t.Menu; t.MenuOpen {
			t.Items, t.Drop = m.treeMenu(r)
		}
		out = append(out, t)
	}
	return out
}

func (m Model) PathLabel() string {
	p := m.OpenPath()
	if m.Scene == "" {
		if n := m.StepShown(); n > 0 {
			return "step " + m.StepLabel(n)
		}
		return ""
	}
	labels := make([]string, len(p))
	for i := range p {
		labels[i] = stepLabelAt(m.pathDir(p[:i]), p[i])
	}
	return strings.Join(labels, " › ")
}

func (m *Model) FindStep(e Sheet.Edit, key string) {
	switch {
	case e.Active:
		m.TreeFind = e.Draft
	case key == "Escape":
		m.TreeFind = ""
	}
	m.ReadTree()
}

func (m Model) treeData(d *LinksBoard.Data) {
	if m.Scene == "" {
		return
	}
	d.Tree = m.treeRows()
	d.TreeFind = &Board.Field{Value: m.TreeFind, Fit: "find a step……", Act: Act{Op: OpTreeFind}, Words: true}
	if len(d.Tree) == 0 && strings.TrimSpace(m.TreeFind) != "" {
		d.TreeEmpty = "no step name contains \"" + m.TreeFind + "\""
	}
}

func (m Model) findsIn(p []int) bool {
	find := strings.ToLower(strings.TrimSpace(m.TreeFind))
	if strings.Contains(strings.ToLower(stepLabelAt(m.pathDir(p[:len(p)-1]), p[len(p)-1])), find) {
		return true
	}
	for _, n := range ReadStepNumbers(m.pathDir(p)) {
		if m.findsIn(append(slices.Clone(p), n)) {
			return true
		}
	}
	return false
}

func (m *Model) walkFind(parent, here []int) {
	dir := m.pathDir(parent)
	for _, n := range ReadStepNumbers(dir) {
		p := append(slices.Clone(parent), n)
		if !m.findsIn(p) {
			continue
		}
		r := StepRow{Path: p, Label: stepLabelAt(dir, n), Depth: len(parent), Shown: slices.Equal(p, here)}
		r.Here = len(p) <= len(here) && slices.Equal(p, here[:len(p)])
		r.Kids = len(ReadStepNumbers(stepDir(dir, n))) > 0
		r.Open = r.Kids
		m.Tree = append(m.Tree, r)
		m.walkFind(p, here)
	}
}
