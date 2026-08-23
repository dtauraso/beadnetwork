package Tabs

import "github.com/dtauraso/wirefold/Categories/Scene/Scenes"

func TabNames() []string {
	names := make([]string, len(Scenes.All))
	for i, s := range Scenes.All {
		names[i] = s.Name
	}
	return names
}

func SelectedIndex(anchorPath string) int {
	return Scenes.SelectedIndex(anchorPath)
}

type State struct {
	Names    []string
	Selected int
}
