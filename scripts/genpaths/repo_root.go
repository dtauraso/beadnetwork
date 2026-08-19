package genpaths

import (
	"os"
	"path/filepath"
)

func FindRepoRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "nodes")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
