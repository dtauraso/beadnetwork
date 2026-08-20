package scenepersist

import (
	"github.com/dtauraso/wirefold/src/valuefile"
	"github.com/dtauraso/wirefold/src/Scene/scene"
	"github.com/dtauraso/wirefold/src/Scene/scenepaths"
)

func WriteSelectedScene(anchorPath string, idx int) error {
	return valuefile.WriteAtomic(scenepaths.SelectionFilePath(anchorPath), scene.All[idx].Name)
}
