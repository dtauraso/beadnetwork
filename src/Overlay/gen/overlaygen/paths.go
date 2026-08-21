package overlaygen

import (
	"os"
	"path/filepath"
)

const RelDir = "view/overlays"

func WritePathFiles(pathsDir string, flags []overlayFlag) error {
	if err := os.MkdirAll(pathsDir, 0o755); err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, f := range flags {
		name := f.flag + ".bin"
		keep[name] = true
		rel := RelDir + "/" + f.flag + ".bin"
		if err := os.WriteFile(filepath.Join(pathsDir, name), []byte(rel), 0o644); err != nil {
			return err
		}
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
