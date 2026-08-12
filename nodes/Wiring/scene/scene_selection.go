package scene

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
)

type SceneSelectionFile struct {
	Selected string `json:"selected"`
}

func AnchorIsTabbed(anchorPath string) bool {
	return filepath.Base(filepath.Clean(anchorPath)) == SceneTabs[0].Dir
}

func SelectedSceneIndex(anchorPath string) int {
	if !AnchorIsTabbed(anchorPath) {
		return 0
	}
	b, err := os.ReadFile(scenepaths.SelectionFilePath(anchorPath))
	if err != nil {
		return 0
	}
	var f SceneSelectionFile
	if json.Unmarshal(b, &f) != nil {
		return 0
	}
	for i, t := range SceneTabs {
		if t.Name == f.Selected {
			return i
		}
	}
	return 0
}

func ResolveScenePath(anchorPath string) string {
	if !AnchorIsTabbed(anchorPath) {
		return anchorPath
	}
	tab := SceneTabs[SelectedSceneIndex(anchorPath)]
	return filepath.Join(filepath.Dir(filepath.Clean(anchorPath)), tab.Dir)
}

func SceneTabNames(anchorPath string) []string {
	if !AnchorIsTabbed(anchorPath) {
		return nil
	}
	names := make([]string, len(SceneTabs))
	for i, t := range SceneTabs {
		names[i] = t.Name
	}
	return names
}
