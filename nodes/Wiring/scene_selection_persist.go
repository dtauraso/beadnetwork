// scene_selection_persist.go — persist Go's OWN scene-tab selection to view/scene.json,
// mirroring scene_camera_persist.go / scene_overlays_persist.go / scene_sphere_persist.go.
//
// OWNER: the view-owner goroutine (RunStdinReader, stdin_reader.go). Its sole trigger is a
// tab click landing as edit-update kind="scene" — applyUpdateScene → MoveDispatch.
// SelectScene → here — so this file has exactly one writer on one goroutine, which is why
// it is a view-owner file in check-persist-write-ownership.sh's list.
//
// WHOLE-FILE write, no read-modify-write: scene.json holds only the selection.
//
// No debounce, unlike the camera's pose. The selection changes at most once per click and
// the very next thing that happens is the process ending (scene_tabs.go's SelectScene), so
// a deferred write would race the exit it precedes — the one case where writing
// immediately is not merely acceptable but required.
//
// The READ side lives in scene_tabs.go (SelectedSceneIndex), next to the tab list it
// resolves against, since resolving a name to an index is tab-list knowledge rather than
// persistence plumbing.
package Wiring

// writeSelectedScene writes the selected tab's NAME (not its index) as the whole content of
// the anchor's view/scene.json. The name is what survives a reordering of SceneTabs; an
// index would silently come back pointing at a different diagram.
func writeSelectedScene(anchorPath string, idx int) error {
	return writeJSONAtomic(sceneSelectionFilePath(anchorPath), sceneSelectionFile{Selected: SceneTabs[idx].Name})
}
