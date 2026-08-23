package scenepersist

import (
	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"
)

func WriteSelectedScene(anchorPath string, idx int) error {
	return WriteAtomic(Scenes.SelectionFilePath(anchorPath), Scenes.All[idx].Name)
}
