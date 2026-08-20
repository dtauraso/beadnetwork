package scenepersist

import (
	"github.com/dtauraso/wirefold/src/jsonpersist"
	"github.com/dtauraso/wirefold/src/Scene/scene"
	"github.com/dtauraso/wirefold/src/Scene/scenepaths"
)

func WriteSelectedScene(anchorPath string, idx int) error {
	return jsonpersist.WriteJSONAtomic(scenepaths.SelectionFilePath(anchorPath), scene.All[idx].Name)
}
