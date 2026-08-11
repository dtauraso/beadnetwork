// scene_switch.go — SelectScene, the *MoveDispatch method that CARRIES OUT a tab-click.
// The tab registry lives in scene/scene_tabs.go; anchor/path resolution against the
// persisted selection lives in scene_selection.go. Arming the switch (formerly
// EnableSceneSwitch, a pure two-field forward onto md.Scenes) was deleted — md.Scenes is
// exported, so its one caller (runtopology/topology_run.go) sets
// md.Scenes.AnchorPath/md.Scenes.Quit directly, before the stdin reader starts — the
// view-owner goroutine is the only caller of SelectScene, so those fields are written
// once, before that goroutine exists.
package Wiring

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
)

// SelectScene handles a tab click: persist, then end the run so the respawn loads it.
// Selecting the tab already showing is a no-op — restarting the sim to arrive at the same
// diagram would look like a random flicker to whoever clicked.
func SelectScene(scenes *sceneswitch.SceneSwitch, idx int) {
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
