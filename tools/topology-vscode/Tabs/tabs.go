package Tabs

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
)

func TabNames() []string {
	names := make([]string, len(scene.All))
	for i, s := range scene.All {
		names[i] = s.Name
	}
	return names
}

func SelectedIndex(anchorPath string) int {
	return scene.SelectedIndex(anchorPath)
}
