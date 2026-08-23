package Scenes

import (
	"fmt"
	"os"
)

func SelectScene(scenes *SceneSwitch, idx int) {
	if scenes.AnchorPath == "" || scenes.Quit == nil {
		return
	}
	if idx < 0 || idx >= len(All) {
		return
	}
	if idx == scenes.Loaded {
		return
	}
	if err := WriteSelectedScene(scenes.AnchorPath, idx); err != nil {

		fmt.Fprintf(os.Stderr, "scene tab: could not persist selection to %s: %v — staying on the current scene\n",
			SelectionFilePath(scenes.AnchorPath), err)
		return
	}
	scenes.Quit()
}
