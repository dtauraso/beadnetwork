package main

import (
	"os"
	"path/filepath"
)

const RelFile = "view/camera.bin"

func writeCameraPaths(cameraDir string) error {
	pathsDir := filepath.Join(cameraDir, "paths")
	if err := os.MkdirAll(pathsDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(pathsDir, "block.bin"), []byte(RelFile), 0o644); err != nil {
		return err
	}

	entries, err := os.ReadDir(pathsDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || e.Name() == "block.bin" { // path-resolution-ok: the generator's own output directory
			continue
		}
		if err := os.Remove(filepath.Join(pathsDir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
