package Scenes

// Which scene is selected, written where the list of scenes is. It used to sit
// with the persistence machinery, which meant the list could not write its own
// selection without an import cycle — SelectScene had to be handed the write.
// It is one line about this list, so it lives with it.
func WriteSelectedScene(anchorPath string, idx int) error {
	return WriteAtomic(SelectionFilePath(anchorPath), All[idx].Name)
}
