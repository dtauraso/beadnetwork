package Builder

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	DirStep       = "step"
	FileLockPhi   = "lock-phi.bin"
	FileLockTheta = "lock-theta.bin"
	FileLocking   = "locking.bin"
)

type Step struct {
	LockAll Links.Lock
	Locking bool
	Round   int
}

func StepDir(nodesRoot string) string { return filepath.Join(nodesRoot, DirStep) }

func ReadStep(dir string) Step {
	var s Step
	Links.ReadIfExists(filepath.Join(dir, FileLockPhi), &s.LockAll.Phi)
	Links.ReadIfExists(filepath.Join(dir, FileLockTheta), &s.LockAll.Theta)
	Links.ReadIfExists(filepath.Join(dir, FileLocking), &s.Locking)
	return s
}

func WriteStep(dir string, s Step) error {
	for name, v := range map[string]any{FileLockPhi: s.LockAll.Phi, FileLockTheta: s.LockAll.Theta, FileLocking: s.Locking} {
		if err := Links.WriteAtomicIfChanged(filepath.Join(dir, name), v); err != nil {
			return err
		}
	}
	return nil
}
