package Builder

import (
	"fmt"
	"os"
	"slices"
)

const (
	OpStepAdd  = "step-add"
	OpStepCopy = "step-copy"
	OpStepDrop = "step-drop"
	OpStepUp   = "step-up"
	OpStepDown = "step-down"
)

func IsStepDirOp(a Act) bool {
	switch a.Op {
	case OpStepAdd, OpStepCopy, OpStepDrop, OpStepUp, OpStepDown:
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
	a, b := stepDir(m.Root, n), stepDir(m.Root, k)
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
	ns := ReadStepNumbers(m.Root)
	n := a.I
	if !slices.Contains(ns, n) {
		return fmt.Sprintf("step %d: there is no steps/%d folder, so nothing changed", n, n), nil
	}
	var err error
	switch a.Op {
	case OpStepAdd:
		if err = m.shiftSteps(ns, n, 1); err == nil {
			err = os.MkdirAll(stepDir(m.Root, n+1), 0o755)
		}
		m.ShownStep = n + 1
	case OpStepCopy:
		if err = m.shiftSteps(ns, n, 1); err == nil {
			err = os.CopyFS(stepDir(m.Root, n+1), os.DirFS(stepDir(m.Root, n)))
		}
		m.ShownStep = n + 1
	case OpStepDrop:
		if len(ns) == 1 {
			return fmt.Sprintf("step %d is the only step, so it stays", n), nil
		}
		if err = os.RemoveAll(stepDir(m.Root, n)); err == nil {
			err = m.shiftSteps(ns, n, -1)
		}
		m.ShownStep = n
	case OpStepUp, OpStepDown:
		var refused string
		if m.ShownStep, refused, err = m.swapSteps(ns, n, map[string]int{OpStepUp: -1, OpStepDown: 1}[a.Op]); refused != "" {
			return refused, nil
		}
	}
	if err != nil {
		return "", err
	}
	m.StepHeads, m.StepFolds = nil, nil
	m.ReadNodeSteps()
	if m.SeedStepLinks() {
		return "", m.WriteSteps()
	}
	return "", nil
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
		if err := os.Rename(stepDir(m.Root, k), stepDir(m.Root, k+by)); err != nil {
			return err
		}
	}
	return nil
}
