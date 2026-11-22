package Builder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/StepStore"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

const (
	OpTreeGo     = "tree-go"
	OpTreeFold   = "tree-fold"
	FileTreeOpen = "tree-open.bin"
)

type StepRow struct {
	Path  []int
	Label string
	Depth int
	Kids  int
	Open  bool
	Here  bool
	Shown bool
}

func (m Model) pathDir(p []int) string {
	dir := m.Scene
	for _, n := range p {
		dir = StepStore.StepDir(dir, n)
	}
	return dir
}

func (m Model) innerPath() []int {
	p := StepStore.LevelPath(m.Level)
	if n := m.StepShown(); n > 0 {
		p = append(p, n)
	}
	return p
}

func (m Model) OpenPath() []int {
	if m.OpenTop > 0 {
		return append([]int{m.OpenTop}, m.innerPath()...)
	}
	return m.innerPath()
}

func (m *Model) ReadTree() {
	m.Tree = nil
	if m.Scene != "" && strings.TrimSpace(m.TreeFind) != "" {
		m.walkFind(nil, m.OpenPath())
	} else if m.Scene != "" {
		m.walkTree(nil, m.OpenPath())
	}
}

func (m *Model) walkTree(parent, here []int) {
	dir := m.pathDir(parent)
	for _, n := range StepStore.ReadStepNumbers(dir) {
		p := append(slices.Clone(parent), n)
		r := StepRow{Path: p, Label: StepStore.LabelAt(dir, n), Depth: len(parent), Shown: slices.Equal(p, here)}
		r.Here = len(p) <= len(here) && slices.Equal(p, here[:len(p)])
		r.Kids = len(StepStore.ReadStepNumbers(StepStore.StepDir(dir, n)))
		r.Open = r.Here && !r.Shown
		Store.ReadIfExists(filepath.Join(StepStore.StepDir(dir, n), FileTreeOpen), &r.Open)
		r.Open = r.Open && r.Kids > 0
		m.Tree = append(m.Tree, r)
		if r.Open {
			m.walkTree(p, here)
		}
	}
}

func (m *Model) ToggleTreeFold(p []int) error {
	for _, r := range m.Tree {
		if slices.Equal(r.Path, p) {
			if err := Store.WriteAtomic(filepath.Join(m.pathDir(p), FileTreeOpen), !r.Open); err != nil {
				return err
			}
		}
	}
	m.ReadTree()
	return nil
}

const (
	OpTreeSplit = "tree-split"
	OpStepName  = "step-name"
)

func (m *Model) StartSplit(left int) { m.Splitting, m.splitX0 = true, float32(left) }

func (m *Model) SplitTo(x float64) bool {
	if !m.Splitting {
		return false
	}
	m.TreeW = max(LinksBoard.TreeMinW, float32(x)-m.splitX0)
	return true
}

func (m *Model) EndSplit() bool {
	was := m.Splitting
	m.Splitting = false
	return was
}

func (m *Model) TreeWheel(pl Sheet.Layout, x, y, dy float64) bool {
	for _, p := range pl.Pills {
		l, ok := p.Board.(*Board.Laid)
		if !ok {
			continue
		}
		for _, s := range l.Spots {
			if v := map[string]*float32{OpTreeScroll: &m.TreeSY}[s.Cell.Act.Op]; v != nil && x >= float64(s.R.X) && x < float64(s.R.X+s.R.W) && y >= float64(s.R.Y) && y < float64(s.R.Y+s.R.H) {
				*v = min(max(0, *v+float32(dy)), float32(s.Cell.Act.I))
				return true
			}
		}
	}
	return false
}

func (m Model) Crosses(p []int) bool { return m.OpenTop > 0 && (len(p) == 0 || p[0] != m.OpenTop) }

func (m *Model) TreeGo(p []int) error {
	inner := p
	if m.OpenTop > 0 {
		inner = p[1:]
	}
	if len(inner) == 0 {
		inner = Scenes.OpenStepPath(m.Root)
	}
	if _, err := os.Stat(filepath.Join(m.Root, StepStore.LevelOf(inner))); err != nil || len(inner) == 0 {
		inner = []int{1}
	}
	if err := m.WriteSteps(); err != nil {
		return err
	}
	m.Level, m.ShownStep = StepStore.LevelOf(inner[:len(inner)-1]), inner[len(inner)-1]
	m.StepClosed = false
	err := m.rereadSteps()
	return err
}

func (m Model) SaveOpen() error {
	if m.Scene == "" || m.Root == "" {
		return nil
	}
	err := Scenes.WriteOpenStep(m.Scene, m.OpenPath())
	if m.Root != m.Scene {
		err = errors.Join(err, Scenes.WriteOpenStep(m.Root, m.innerPath()))
	}
	return err
}

const (
	OpStepAdd     = "step-add"
	OpStepCopy    = "step-copy"
	OpStepDrop    = "step-drop"
	OpStepEarlier = "step-earlier"
	OpStepLater   = "step-later"
)

func IsStepDirOp(a Act) bool {
	switch a.Op {
	case OpStepAdd, OpStepCopy, OpStepDrop, OpStepEarlier, OpStepLater:
		return true
	}
	return false
}

func (m *Model) StepDirOp(a Act) (string, error) {
	if m.Root == "" {
		return "steps are kept in the scene's steps/ folder, and this scene has none, so nothing changed", nil
	}
	if err := m.WriteSteps(); err != nil {
		return "", err
	}
	ns := StepStore.ReadStepNumbers(m.Lvl())
	n := a.I
	if !slices.Contains(ns, n) {
		return fmt.Sprintf("step %d: there is no steps/%d folder, so nothing changed", n, n), nil
	}
	if by := map[string]int{OpStepEarlier: -1, OpStepLater: 1}[a.Op]; by != 0 {
		i := slices.Index(ns, n) + by
		if i < 0 || i >= len(ns) {
			return fmt.Sprintf("step %d is already the %s step, so it stays", n, map[int]string{-1: "first", 1: "last"}[by]), nil
		}
		return m.MoveStep(n, ns[i])
	}
	var err error
	m.StepFrom = shiftedFrom(ns, n, a.Op)
	switch a.Op {
	case OpStepAdd:
		if err = m.ShiftSteps(ns, n, 1); err == nil {
			err = os.MkdirAll(StepStore.StepDir(m.Lvl(), n+1), 0o755)
		}
		m.ShownStep = n + 1
		m.setStepOpen(n+1, true)
	case OpStepCopy:
		if err = m.ShiftSteps(ns, n, 1); err == nil {
			err = os.CopyFS(StepStore.StepDir(m.Lvl(), n+1), os.DirFS(StepStore.StepDir(m.Lvl(), n)))
		}
		m.ShownStep = n + 1
		m.setStepOpen(n+1, true)
	case OpStepDrop:
		if len(ns) == 1 {
			return fmt.Sprintf("step %d is the only step, so it stays", n), nil
		}
		if err = os.RemoveAll(StepStore.StepDir(m.Lvl(), n)); err == nil {
			err = m.ShiftSteps(ns, n, -1)
		}
		m.ShownStep = n
	}
	if err != nil {
		return "", err
	}
	return "", m.rereadSteps()
}

func (m *Model) rereadSteps() error {
	m.Heads = nil
	m.ReadNodeSteps()
	m.ReadTree()
	return nil
}

func shiftedFrom(ns []int, n int, op string) []int {
	var from []int
	for _, k := range ns {
		switch {
		case k < n || (k == n && op != OpStepDrop):
			from = append(from, k)
		case k > n && op == OpStepDrop:
			from = append(from, k)
		}
		if k == n && op == OpStepAdd {
			from = append(from, 0)
		}
		if k == n && op == OpStepCopy {
			from = append(from, n)
		}
	}
	return from
}

func movedFrom(ns []int, from, to int) []int {
	order := slices.DeleteFunc(slices.Clone(ns), func(k int) bool { return k == from })
	return slices.Insert(order, slices.Index(ns, to), from)
}

func (m *Model) MoveStep(from, to int) (string, error) {
	if m.Root == "" {
		return "steps are kept in the scene's steps/ folder, and this scene has none, so nothing moved", nil
	}
	if err := m.WriteSteps(); err != nil {
		return "", err
	}
	ns := StepStore.ReadStepNumbers(m.Lvl())
	if !slices.Contains(ns, from) || !slices.Contains(ns, to) {
		return fmt.Sprintf("step %d or step %d has no folder, so nothing moved", from, to), nil
	}
	by := 1
	if to < from {
		by = -1
	}
	for at := from; at != to; {
		next, refused, err := m.SwapSteps(ns, at, by)
		if refused != "" || err != nil {
			return refused, err
		}
		at = next
	}
	m.ShownStep, m.StepClosed, m.StepFrom = to, false, movedFrom(ns, from, to)
	return "", m.rereadSteps()
}

func (m *Model) DropLevel(p []int) (bool, error) {
	inner := p
	if m.OpenTop > 0 && len(p) > 0 {
		inner = p[1:]
	}
	up := p[:max(0, len(p)-1)]
	if len(inner) == 0 || (len(inner) == 1 && m.OpenTop == 0) || len(StepStore.ReadStepNumbers(m.pathDir(up))) != 1 {
		return false, nil
	}
	if len(inner) == 1 {
		if err := m.WriteSteps(); err != nil {
			return true, err
		}
		m.Level, m.ShownStep, m.StepClosed = "", 1, false
	} else if err := m.TreeGo(up); err != nil {
		return true, err
	}
	if err := os.RemoveAll(StepStore.StepsDir(m.pathDir(up))); err != nil {
		return true, err
	}
	return true, m.rereadSteps()
}
