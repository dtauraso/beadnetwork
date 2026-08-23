package Scenes

import (
	"fmt"
	"os"
)

func SelectScene(scenes *SceneSwitch, idx int, writeSelected func(anchorPath string, idx int) error) {
	if scenes.AnchorPath == "" || scenes.Quit == nil {
		return
	}
	if idx < 0 || idx >= len(All) {
		return
	}
	if idx == scenes.Loaded {
		return
	}
	if err := writeSelected(scenes.AnchorPath, idx); err != nil {

		fmt.Fprintf(os.Stderr, "scene tab: could not persist selection to %s: %v — staying on the current scene\n",
			SelectionFilePath(scenes.AnchorPath), err)
		return
	}
	scenes.Quit()
}
