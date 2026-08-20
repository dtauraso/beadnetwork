package Camera

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const PathsDirName = "paths"

type Paths map[string]string

func LoadPaths(cameraSrcDir string) (Paths, error) {
	dir := filepath.Join(cameraSrcDir, PathsDirName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := Paths{}
	for _, e := range entries {

		if e.IsDir() || !strings.HasSuffix(e.Name(), valuefile.Ext) { // path-resolution-ok: the camera's own paths dir, not a scene path
			continue
		}
		var rel string
		if !valuefile.ReadIfExists(filepath.Join(dir, e.Name()), &rel) {
			continue
		}
		out[strings.TrimSuffix(e.Name(), valuefile.Ext)] = rel
	}
	return out, nil
}

func (p Paths) FileFor(sceneRoot, primitive string) (string, bool) {
	rel, ok := p[primitive]
	if !ok {
		return "", false
	}
	return filepath.Join(sceneRoot, rel), true
}
