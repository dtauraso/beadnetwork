// scene_switch.go — the two *MoveDispatch methods that PERFORM a tab switch: arming it
// (EnableSceneSwitch) and carrying out a click (SelectScene). The tab registry lives in
// scene/scene_tabs.go; anchor/path resolution against the persisted selection lives in
// scene_selection.go.
package Wiring

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

// EnableSceneSwitch arms tab switching. quit ends the run (main's context cancel), which
// the extension host's looping runner follows with a respawn. Called from main.go after
// load, before the stdin reader starts — the view-owner goroutine is the only caller of
// SelectScene, so this field is written once, before that goroutine exists.
func (md *MoveDispatch) EnableSceneSwitch(anchorPath string, quit func()) {
	md.Scenes.AnchorPath = anchorPath
	md.Scenes.Quit = quit
}

// SelectScene handles a tab click: persist, then end the run so the respawn loads it.
// Selecting the tab already showing is a no-op — restarting the sim to arrive at the same
// diagram would look like a random flicker to whoever clicked.
func (md *MoveDispatch) SelectScene(idx int) {
	if md.Scenes.AnchorPath == "" || md.Scenes.Quit == nil {
		return
	}
	if idx < 0 || idx >= len(scene.SceneTabs) {
		return
	}
	if idx == scene.SelectedSceneIndex(md.Scenes.AnchorPath) {
		return
	}
	if err := scenepersist.WriteSelectedScene(md.Scenes.AnchorPath, idx); err != nil {
		// Do NOT quit on a failed write: the respawn would reload the OLD scene and the
		// click would read as "the editor restarted for no reason". Report and stay put.
		// stderr, not a breadcrumb: the extension host pipes this straight to the sim's
		// output channel AND .probe/go-errors.jsonl, which is where an operator looks when
		// a click did nothing (memory/feedback_runner_errors_probe_first.md).
		fmt.Fprintf(os.Stderr, "scene tab: could not persist selection to %s: %v — staying on the current scene\n",
			scenepaths.SelectionFilePath(md.Scenes.AnchorPath), err)
		return
	}
	md.Scenes.Quit()
}
