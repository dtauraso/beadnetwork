package Builder

import (
	"fmt"
	"os"
	"slices"
)

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

func (m Model) swapSteps(ns []int, n, by int) (int, string, error) {
	i := slices.Index(ns, n) + by
	if i < 0 || i >= len(ns) {
		return n, fmt.Sprintf("step %d is already the %s step, so it stays", n, map[int]string{-1: "first", 1: "last"}[by]), nil
	}
	k := ns[i]
	a, b := stepDir(m.lvl(), n), stepDir(m.lvl(), k)
	swap := a + ".swap"
	if err := os.Rename(a, swap); err != nil {
		return n, "", err
	}
	if err := os.Rename(b, a); err != nil {
		return n, "", err
	}
	return k, "", os.Rename(swap, b)
}

func (m *Model) StepDirOp(a Act) (string, error) {
	if m.Root == "" {
		return "steps are kept in the scene's steps/ folder, and this scene has none, so nothing changed", nil
	}
	if err := m.WriteSteps(); err != nil {
		return "", err
	}
	ns := ReadStepNumbers(m.lvl())
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
		if err = m.shiftSteps(ns, n, 1); err == nil {
			err = os.MkdirAll(stepDir(m.lvl(), n+1), 0o755)
		}
		m.ShownStep = n + 1
		m.setStepOpen(n+1, true)
	case OpStepCopy:
		if err = m.shiftSteps(ns, n, 1); err == nil {
			err = os.CopyFS(stepDir(m.lvl(), n+1), os.DirFS(stepDir(m.lvl(), n)))
		}
		m.ShownStep = n + 1
		m.setStepOpen(n+1, true)
	case OpStepDrop:
		if len(ns) == 1 {
			return fmt.Sprintf("step %d is the only step, so it stays", n), nil
		}
		if err = os.RemoveAll(stepDir(m.lvl(), n)); err == nil {
			err = m.shiftSteps(ns, n, -1)
		}
		m.ShownStep = n
	}
	if err != nil {
		return "", err
	}
	return "", m.rereadSteps()
}

func (m *Model) rereadSteps() error {
	m.StepHeads, m.StepFolds, m.Touched = nil, nil, StepEdge{}
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
	ns := ReadStepNumbers(m.lvl())
	if !slices.Contains(ns, from) || !slices.Contains(ns, to) {
		return fmt.Sprintf("step %d or step %d has no folder, so nothing moved", from, to), nil
	}
	by := 1
	if to < from {
		by = -1
	}
	for at := from; at != to; {
		next, refused, err := m.swapSteps(ns, at, by)
		if refused != "" || err != nil {
			return refused, err
		}
		at = next
	}
	m.ShownStep, m.StepClosed, m.StepFrom = to, false, movedFrom(ns, from, to)
	return "", m.rereadSteps()
}

func (m Model) shiftSteps(ns []int, after, by int) error {
	var later []int
	for _, k := range ns {
		if k > after {
			later = append(later, k)
		}
	}
	if by > 0 {
		slices.Reverse(later)
	}
	for _, k := range later {
		if err := os.Rename(stepDir(m.lvl(), k), stepDir(m.lvl(), k+by)); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) DropLevel(p []int) (bool, error) {
	inner := p
	if m.OpenTop > 0 && len(p) > 0 {
		inner = p[1:]
	}
	up := p[:max(0, len(p)-1)]
	if len(inner) < 2 || len(ReadStepNumbers(m.pathDir(up))) != 1 {
		return false, nil
	}
	if err := m.TreeGo(up); err != nil {
		return true, err
	}
	if err := os.RemoveAll(StepsDir(m.pathDir(up))); err != nil {
		return true, err
	}
	return true, m.rereadSteps()
}
