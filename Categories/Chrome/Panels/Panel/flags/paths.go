package main

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/scripts/genpaths"
)

const RelFile = "view/panels.bin"

func WritePathFiles(pathsDir string, flags []string) error {
	if err := os.MkdirAll(pathsDir, 0o755); err != nil {
		return err
	}
	keep := map[string]bool{"block.bin": true}
	if err := genpaths.WriteIfChanged(filepath.Join(pathsDir, "block.bin"), []byte(RelFile), 0o644); err != nil {
		return err
	}

	if err := writeGoPaths(filepath.Dir(pathsDir), flags); err != nil {
		return err
	}

	entries, err := os.ReadDir(pathsDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || keep[e.Name()] { // path-resolution-ok: the generator's own output directory
			continue
		}
		if err := os.Remove(filepath.Join(pathsDir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
