package Builder

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	DirPress        = "press"
	DirSteps        = "steps"
	DirNodes        = "nodes"
	FileLockPhi     = "lock-phi.bin"
	FileLockTheta   = "lock-theta.bin"
	FileLockR       = "lock-r.bin"
	FileLockOnPhi   = "lock-on-phi.bin"
	FileLockOnTheta = "lock-on-theta.bin"
	FileLockOnR     = "lock-on-r.bin"
	FileLockOnBoth  = "lock-on-both.bin"
)

type LockOn struct{ Phi, Theta, R bool }

type Step struct {
	LockAll Links.Lock
	On      LockOn
	Both    bool
	Round   int
}

func (s Step) Locks() Links.Lock {
	l := s.LockAll
	if !s.On.Phi {
		l.Phi = 0
	}
	if !s.On.Theta {
		l.Theta = 0
	}
	if !s.On.R {
		l.R = 0
	}
	return l
}

func (s *Step) ToggleBoth() {
	s.Both = !s.Both
	if s.Both {
		s.On = LockOn{}
	}
}

func (s *Step) At(k string) (*int, *bool) {
	switch k {
	case Links.KTheta:
		return &s.LockAll.Theta, &s.On.Theta
	case Links.KR:
		return &s.LockAll.R, &s.On.R
	}
	return &s.LockAll.Phi, &s.On.Phi
}

func PressDir(root string) string { return filepath.Join(root, DirPress) }

func nodePressDir(root, id string) string { return filepath.Join(root, DirNodes, id, DirPress) }

func StepsDir(root string) string { return filepath.Join(root, DirSteps) }

func stepDir(root string, step int) string { return filepath.Join(StepsDir(root), strconv.Itoa(step)) }

func stepNodeDir(root string, step int, id string) string {
	return filepath.Join(stepDir(root, step), DirNodes, id)
}

func ReadStepNumbers(root string) []int {
	var out []int
	entries, _ := os.ReadDir(StepsDir(root))
	for _, e := range entries {
		if n, err := strconv.Atoi(e.Name()); e.IsDir() && err == nil && n > 0 { // path-resolution-ok: steps/ holds one numbered directory per step; skipping strays
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out
}

func (m Model) stepNums() []int {
	if len(m.StepNums) == 0 {
		return []int{1}
	}
	return m.StepNums
}

func (m Model) StepOf(id string) Step {
	if id == "" {
		return m.Step
	}
	if s := m.Steps[id]; s != nil {
		return *s
	}
	return Step{}
}

func (m *Model) NodeStep(id string) *Step {
	if id == "" {
		return &m.Step
	}
	if m.Steps == nil {
		m.Steps = map[string]*Step{}
	}
	if m.Steps[id] == nil {
		s := Step{}
		if m.Root != "" {
			s = ReadStep(nodePressDir(m.Root, id))
		}
		m.Steps[id] = &s
	}
	return m.Steps[id]
}

func (m *Model) ReadNodeSteps() {
	for _, id := range m.Nodes {
		m.NodeStep(id)
	}
	if m.Root != "" {
		m.StepNums = ReadStepNumbers(m.Root)
		m.readStepLinks()
		for _, step := range m.stepNums() {
			m.StepHead(step)
		}
	}
}

func (m Model) WriteSteps() error {
	if err := WriteStep(PressDir(m.Root), m.Step); err != nil {
		return err
	}
	for id, s := range m.Steps {
		if err := WriteStep(nodePressDir(m.Root, id), *s); err != nil {
			return err
		}
	}
	if err := m.writeStepLinks(); err != nil {
		return err
	}
	return m.writeStepHeads()
}

func ReadStep(dir string) Step {
	var s Step
	Links.ReadIfExists(filepath.Join(dir, FileLockPhi), &s.LockAll.Phi)
	Links.ReadIfExists(filepath.Join(dir, FileLockTheta), &s.LockAll.Theta)
	Links.ReadIfExists(filepath.Join(dir, FileLockR), &s.LockAll.R)
	s.On = LockOn{s.LockAll.Phi != 0, s.LockAll.Theta != 0, s.LockAll.R != 0}
	Links.ReadIfExists(filepath.Join(dir, FileLockOnPhi), &s.On.Phi)
	Links.ReadIfExists(filepath.Join(dir, FileLockOnTheta), &s.On.Theta)
	Links.ReadIfExists(filepath.Join(dir, FileLockOnR), &s.On.R)
	Links.ReadIfExists(filepath.Join(dir, FileLockOnBoth), &s.Both)
	return s
}

func WriteStep(dir string, s Step) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, v := range map[string]any{FileLockPhi: s.LockAll.Phi, FileLockTheta: s.LockAll.Theta, FileLockR: s.LockAll.R,
		FileLockOnPhi: s.On.Phi, FileLockOnTheta: s.On.Theta, FileLockOnR: s.On.R, FileLockOnBoth: s.Both} {
		if err := Links.WriteAtomicIfChanged(filepath.Join(dir, name), v); err != nil {
			return err
		}
	}
	return nil
}
