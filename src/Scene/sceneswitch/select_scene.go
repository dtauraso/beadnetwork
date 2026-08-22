package sceneswitch

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/src/Scene/scene"

	"github.com/dtauraso/wirefold/src/Scene/scenepaths"
	"github.com/dtauraso/wirefold/src/Scene/scenepersist"
)

func SelectScene(scenes *SceneSwitch, idx int) {
	if scenes.AnchorPath == "" || scenes.Quit == nil {
		return
	}
	if idx < 0 || idx >= len(scene.All) {
		return
	}
	if idx == scenes.Loaded {
		return
	}
	if err := scenepersist.WriteSelectedScene(scenes.AnchorPath, idx); err != nil {

		fmt.Fprintf(os.Stderr, "scene tab: could not persist selection to %s: %v — staying on the current scene\n",
			scenepaths.SelectionFilePath(scenes.AnchorPath), err)
		return
	}
	scenes.Quit()
}
