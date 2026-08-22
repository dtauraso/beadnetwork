package scenepersist

import (
	"github.com/dtauraso/wirefold/Categories/Scene/scene"
	"github.com/dtauraso/wirefold/Categories/Scene/scenepaths"
)

func WriteSelectedScene(anchorPath string, idx int) error {
	return WriteAtomic(scenepaths.SelectionFilePath(anchorPath), scene.All[idx].Name)
}
