package scenepersist

import (
	"github.com/dtauraso/wirefold/Scenes"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
)

func WriteSelectedScene(anchorPath string, idx int) error {
	return jsonpersist.WriteJSONAtomic(scenepaths.SelectionFilePath(anchorPath), Scenes.SelectionFile{Selected: Scenes.All[idx].Name})
}
