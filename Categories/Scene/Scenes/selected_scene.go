package Scenes

func WriteSelectedScene(anchorPath string, idx int) error {
	return WriteAtomic(SelectionFilePath(anchorPath), All[idx].Name)
}
