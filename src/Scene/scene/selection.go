package scene

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/src/Scene/scenepaths"
	"github.com/dtauraso/wirefold/src/valuefile"
)

func Container(anchorPath string) string {
	clean := filepath.Clean(anchorPath)
	if _, ok := Declared(clean); ok {
		return filepath.Dir(clean)
	}
	return clean
}

func SelectedIndex(anchorPath string) int {
	var selected string
	if !valuefile.ReadIfExists(scenepaths.SelectionFilePath(anchorPath), &selected) {
		return 0
	}
	for i, s := range All {
		if s.Name == selected {
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
