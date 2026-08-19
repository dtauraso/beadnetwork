package scenepersist

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scene"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepaths"
)

func WriteSelectedScene(anchorPath string, idx int) error {
	return jsonpersist.WriteJSONAtomic(scenepaths.SelectionFilePath(anchorPath), scene.SelectionFile{Selected: scene.All[idx].Name})
}
