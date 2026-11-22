package StepStore

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
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
	FilePlayScale   = "play-scale.bin"
	FileStepName    = "name.bin"
)

type LockOn struct{ Phi, Theta, R bool }

type Step struct {
	LockAll Links.Lock
	On      LockOn
	Both    bool
	Round   int
	Scale   int
}

func (s Step) PlayScale() int { return max(1, s.Scale) }

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

func NodePressDir(root, id string) string { return filepath.Join(root, DirNodes, id, DirPress) }

func StepsDir(root string) string { return filepath.Join(root, DirSteps) }

func StepDir(root string, step int) string { return filepath.Join(StepsDir(root), strconv.Itoa(step)) }

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

func ReadStep(dir string) Step {
	var s Step
	Store.ReadIfExists(filepath.Join(dir, FileLockPhi), &s.LockAll.Phi)
	Store.ReadIfExists(filepath.Join(dir, FileLockTheta), &s.LockAll.Theta)
	Store.ReadIfExists(filepath.Join(dir, FileLockR), &s.LockAll.R)
	s.On = LockOn{s.LockAll.Phi != 0, s.LockAll.Theta != 0, s.LockAll.R != 0}
	Store.ReadIfExists(filepath.Join(dir, FileLockOnPhi), &s.On.Phi)
	Store.ReadIfExists(filepath.Join(dir, FileLockOnTheta), &s.On.Theta)
	Store.ReadIfExists(filepath.Join(dir, FileLockOnR), &s.On.R)
	Store.ReadIfExists(filepath.Join(dir, FileLockOnBoth), &s.Both)
	Store.ReadIfExists(filepath.Join(dir, FilePlayScale), &s.Scale)
	return s
}

func WriteStep(dir string, s Step) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, v := range map[string]any{FileLockPhi: s.LockAll.Phi, FileLockTheta: s.LockAll.Theta, FileLockR: s.LockAll.R,
		FileLockOnPhi: s.On.Phi, FileLockOnTheta: s.On.Theta, FileLockOnR: s.On.R, FileLockOnBoth: s.Both, FilePlayScale: s.PlayScale()} {
		if err := Store.WriteAtomicIfChanged(filepath.Join(dir, name), v); err != nil {
			return err
		}
	}
	return nil
}

func LabelAt(dir string, n int) string {
	var name string
	if Store.ReadIfExists(filepath.Join(StepDir(dir, n), FileStepName), &name) && name != "" {
		return "step " + strconv.Itoa(n) + " · " + name
	}
	return "step " + strconv.Itoa(n)
}

func PathKey(p []int) string {
	parts := make([]string, len(p))
	for i, n := range p {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, "/")
}

func ParsePath(s string) []int {
	var p []int
	for _, part := range strings.Split(s, "/") {
		if n, err := strconv.Atoi(part); err == nil && n > 0 {
			p = append(p, n)
		}
	}
	return p
}

func LevelPath(level string) []int {
	var p []int
	parts := strings.Split(filepath.ToSlash(level), "/")
	for i := 1; i < len(parts); i += 2 {
		if n, err := strconv.Atoi(parts[i]); err == nil && parts[i-1] == DirSteps {
			p = append(p, n)
		}
	}
	return p
}

func LevelOf(p []int) string {
	var parts []string
	for _, n := range p {
		parts = append(parts, DirSteps, strconv.Itoa(n))
	}
	return filepath.Join(parts...)
}
