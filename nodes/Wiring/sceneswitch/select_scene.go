package sceneswitch

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

func SelectScene(scenes *SceneSwitch, idx int) {
	if scenes.AnchorPath == "" || scenes.Quit == nil {
		return
	}
	if idx < 0 || idx >= len(scene.Scenes) {
		return
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
