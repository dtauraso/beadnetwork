package sceneswitch

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

// SelectScene handles a tab click: persist, then end the run so the respawn loads it.
// Selecting the tab already showing is a no-op — restarting the sim to arrive at the same
// diagram would look like a random flicker to whoever clicked. Moved from
// nodes/Wiring/dispatch's scene_switch.go (docs/planning/movedispatch-decomposition.md,
// the remainder cluster) — it already took *SceneSwitch as its sole owner parameter, so the
// move is a pure relocation with the caller (nodes/Wiring/stdinreader's applyUpdateScene)
// re-qualified to sceneswitch.SelectScene.
func SelectScene(scenes *SceneSwitch, idx int) {
	if scenes.AnchorPath == "" || scenes.Quit == nil {
		return
	}
	if idx < 0 || idx >= len(scene.SceneTabs) {
		return
	}
	if idx == scene.SelectedSceneIndex(scenes.AnchorPath) {
		return
	}
	if err := scenepersist.WriteSelectedScene(scenes.AnchorPath, idx); err != nil {
		// Do NOT quit on a failed write: the respawn would reload the OLD scene and the
		// click would read as "the editor restarted for no reason". Report and stay put.
		// stderr, not a breadcrumb: the extension host pipes this straight to the sim's
		// output channel AND .probe/go-errors.jsonl, which is where an operator looks when
		// a click did nothing (memory/feedback_runner_errors_probe_first.md).
		fmt.Fprintf(os.Stderr, "scene tab: could not persist selection to %s: %v — staying on the current scene\n",
			scenepaths.SelectionFilePath(scenes.AnchorPath), err)
		return
	}
	scenes.Quit()
}
