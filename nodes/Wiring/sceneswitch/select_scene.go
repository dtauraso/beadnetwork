package sceneswitch

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

func SelectScene(scenes *SceneSwitch, idx int) {
	if scenes.AnchorPath == "" {
		panic("SelectScene: a scene tab was clicked but the SceneSwitch has no AnchorPath, so there is no scene dir to persist the selection into — the switch was never wired to the loaded topology")
	}
	if scenes.Quit == nil {
		panic("SelectScene: a scene tab was clicked but the SceneSwitch has no Quit func, so the selection could be written yet the sim could never reload into the picked scene — the switch was built without its reload hook")
	}
	if idx < 0 || idx >= len(scene.SceneTabs) {
		panic(fmt.Sprintf("SelectScene: tab index %d is outside the %d scene tabs the buffer published, so the webview and scene.SceneTabs disagree about how many tabs exist", idx, len(scene.SceneTabs)))
	}
	if idx == scene.SelectedSceneIndex(scenes.AnchorPath) {
		return
	}
	if err := scenepersist.WriteSelectedScene(scenes.AnchorPath, idx); err != nil {

		fmt.Fprintf(os.Stderr, "scene tab: could not persist selection to %s: %v — staying on the current scene\n",
			scenepaths.SelectionFilePath(scenes.AnchorPath), err)
		return
	}
	scenes.Quit()
}
