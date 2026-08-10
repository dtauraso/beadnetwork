// scene_selection_persist.go — persist Go's OWN scene-tab selection to view/scene.json,
// mirroring scene_camera_persist.go / scene_overlays_persist.go / scene_sphere_persist.go.
//
// OWNER: the view-owner goroutine (RunStdinReader, nodes/Wiring/stdinreader/stdin_reader.go).
// Its sole trigger is a tab click landing as edit-update kind="scene" — applyUpdateScene →
// MoveDispatch.SelectScene → here — so this file has exactly one writer on one goroutine,
// which is why it is a view-owner file in check-persist-write-ownership.sh's list (matched
// by basename, so this package split does not need that guard touched).
//
// WHOLE-FILE write, no read-modify-write: scene.json holds only the selection.
//
// No debounce, unlike the camera's pose. The selection changes at most once per click and
// the very next thing that happens is the process ending (scene_switch.go's SelectScene), so
// a deferred write would race the exit it precedes — the one case where writing
// immediately is not merely acceptable but required.
//
// The READ side lives in scene/scene_selection.go (SelectedSceneIndex), next to the tab list it
// resolves against, since resolving a name to an index is tab-list knowledge rather than
// persistence plumbing.
package scenepersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
)

// WriteSelectedScene writes the selected tab's NAME (not its index) as the whole content of
// the anchor's view/scene.json. The name is what survives a reordering of SceneTabs; an
// index would silently come back pointing at a different diagram.
func WriteSelectedScene(anchorPath string, idx int) error {
	return jsonpersist.WriteJSONAtomic(scenepaths.SelectionFilePath(anchorPath), scene.SceneSelectionFile{Selected: scene.SceneTabs[idx].Name})
}
