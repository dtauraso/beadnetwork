package scene

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
)

type SelectionFile struct {
	Selected string `json:"selected"`
}

func Container(anchorPath string) string {
	clean := filepath.Clean(anchorPath)
	if _, ok := Declared(clean); ok {
		return filepath.Dir(clean)
	}
	return clean
}

func SelectedIndex(anchorPath string) int {
	b, err := os.ReadFile(scenepaths.SelectionFilePath(anchorPath))
	if err != nil {
		return 0
	}
	var f SelectionFile
	if json.Unmarshal(b, &f) != nil {
		return 0
	}
	for i, s := range All {
		if s.Name == f.Selected {
			return i
		}
	}
	return 0
}

func ResolvePath(anchorPath string) string {
	selected := All[SelectedIndex(anchorPath)]
	container := Container(anchorPath)
	resolved := filepath.Join(container, selected.Dir)
	if info, err := os.Stat(resolved); err != nil || !info.IsDir() { // path-resolution-ok: asserting the resolver's own output exists, not resolving a second way
		panic(fmt.Sprintf(
			"ResolveScenePath: scene %q selected at anchor %s resolves to %s, which is not a directory — "+
				"every scene in scene.All must exist beside the anchor, since the tab strip offers all of them",
			selected.Name, anchorPath, resolved))
	}
	return resolved
}
