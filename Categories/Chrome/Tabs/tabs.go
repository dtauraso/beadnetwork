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

	// The strip's own writer, armed when the scene opens. It is unexported:
	// nothing outside can write this block.
	w *ValueWriter
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }
