package Builder

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	DirStep         = "step"
	FileLockPhi     = "lock-phi.bin"
	FileLockTheta   = "lock-theta.bin"
	FileLockR       = "lock-r.bin"
	FileLockOnPhi   = "lock-on-phi.bin"
	FileLockOnTheta = "lock-on-theta.bin"
	FileLockOnR     = "lock-on-r.bin"
	FileLocking     = "locking.bin"
)

type LockOn struct{ Phi, Theta, R bool }

type Step struct {
	LockAll Links.Lock
	On      LockOn
	Locking bool
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

func (s *Step) At(k string) (*int, *bool) {
	switch k {
	case Links.KTheta:
		return &s.LockAll.Theta, &s.On.Theta
	case Links.KR:
		return &s.LockAll.R, &s.On.R
	}
	return &s.LockAll.Phi, &s.On.Phi
}

func StepDir(nodesRoot string) string { return filepath.Join(nodesRoot, DirStep) }

func ReadStep(dir string) Step {
	var s Step
	Links.ReadIfExists(filepath.Join(dir, FileLockPhi), &s.LockAll.Phi)
	Links.ReadIfExists(filepath.Join(dir, FileLockTheta), &s.LockAll.Theta)
	Links.ReadIfExists(filepath.Join(dir, FileLockR), &s.LockAll.R)
	s.On = LockOn{s.LockAll.Phi != 0, s.LockAll.Theta != 0, s.LockAll.R != 0}
	Links.ReadIfExists(filepath.Join(dir, FileLockOnPhi), &s.On.Phi)
	Links.ReadIfExists(filepath.Join(dir, FileLockOnTheta), &s.On.Theta)
	Links.ReadIfExists(filepath.Join(dir, FileLockOnR), &s.On.R)
	Links.ReadIfExists(filepath.Join(dir, FileLocking), &s.Locking)
	return s
}

func WriteStep(dir string, s Step) error {
	for name, v := range map[string]any{FileLockPhi: s.LockAll.Phi, FileLockTheta: s.LockAll.Theta, FileLockR: s.LockAll.R,
		FileLockOnPhi: s.On.Phi, FileLockOnTheta: s.On.Theta, FileLockOnR: s.On.R, FileLocking: s.Locking} {
		if err := Links.WriteAtomicIfChanged(filepath.Join(dir, name), v); err != nil {
			return err
		}
	}
	return nil
}
