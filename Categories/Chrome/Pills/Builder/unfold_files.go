package Builder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const DirUnfold = "unfold"

func UnfoldDir(nodesRoot string) string { return filepath.Join(nodesRoot, DirUnfold) }

func ReadUnfold(dir string) map[string]bool {
	u := map[string]bool{}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		k, ok := strings.CutSuffix(e.Name(), ".bin")
		if !ok || e.IsDir() { // path-resolution-ok: one bool file per foldable column
			continue
		}
		var on bool
		Links.ReadIfExists(filepath.Join(dir, e.Name()), &on)
		u[k] = on
	}
	return u
}

func WriteUnfold(dir string, u map[string]bool) error {
	for k, on := range u {
		if err := Links.WriteAtomicIfChanged(filepath.Join(dir, k+".bin"), on); err != nil {
			return err
		}
	}
	return nil
}
