// scene_selection.go — ANCHOR/PATH RESOLUTION for the scene tabs: which anchor is tabbed,
// which tab is currently selected, and the directory that selection maps to. The tab
// REGISTRY itself (SceneTab, SceneTabs) lives in scene_tabs.go; the switch (the two
// *MoveDispatch methods that perform it) lives in Wiring's scene_switch.go (a MoveDispatch
// method cannot live in this package); the per-scene capability queries derived from a
// resolved scene path live in scene_capabilities.go.
package scene

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
)

// SceneSelectionFile is the persisted selection, held at the ANCHOR (never inside a scene).
type SceneSelectionFile struct {
	Selected string `json:"selected"`
}

// AnchorIsTabbed reports whether this anchor is the tabbed layout's anchor — i.e. whether
// its basename is tab 0's Dir. A test fixture or a one-off tree launched from somewhere
// else is NOT tabbed: it loads exactly itself, streams an empty tab strip, and no tab UI
// appears. That keeps every existing fixture working untouched rather than requiring each
// to grow a scenes layout.
func AnchorIsTabbed(anchorPath string) bool {
	return filepath.Base(filepath.Clean(anchorPath)) == SceneTabs[0].Dir
}

// SelectedSceneIndex reads the persisted selection for this anchor. An absent, malformed,
// or unknown-name file means tab 0 — a fresh checkout has no selection, and that is the
// normal case, not an error.
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

// ResolveScenePath maps an anchor to the directory LoadTopology should actually load: the
// selected tab's sibling directory, or the anchor itself when this anchor is not tabbed.
func ResolveScenePath(anchorPath string) string {
	if !AnchorIsTabbed(anchorPath) {
		return anchorPath
	}
	tab := SceneTabs[SelectedSceneIndex(anchorPath)]
	return filepath.Join(filepath.Dir(filepath.Clean(anchorPath)), tab.Dir)
}

// SceneTabNames is the label list streamed on the VIEW frame. Empty for an untabbed
// anchor, which is what makes the strip absent rather than a single dead tab.
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
