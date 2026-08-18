package scenepersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
)

func WriteSelectedScene(anchorPath string, idx int) error {
	return jsonpersist.WriteJSONAtomic(scenepaths.SelectionFilePath(anchorPath), scene.SelectionFile{Selected: scene.All[idx].Name})
}
