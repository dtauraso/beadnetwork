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

// State is what the strip remembers: the scenes it offers and which is open.
type State struct {
	Names    []string
	Selected int
}
