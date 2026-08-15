package dragfile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

func DragPath(root, id string) string {
	return filepath.Join(root, "nodes", id, "drag", "self.json")
}

type JSON struct {
	DragPolarR     float64 `json:"dragPolarR"`
	DragPolarPhi   float64 `json:"dragPolarPhi"`
	DragPolarTheta float64 `json:"dragPolarTheta"`
	IndexPhi       int     `json:"indexPhi"`
	IndexTheta     int     `json:"indexTheta"`
	IndexR         int     `json:"indexR"`

	TopTiltVectorPhiIdx int32 `json:"topTiltVectorThetaIdx,omitempty"`
}

func Write(root, id string, j JSON) error {
	path := DragPath(root, id)
	return jsonpersist.ReadModifyWriteJSON(path, func(m map[string]any) {
		m["dragPolarR"] = j.DragPolarR
		m["dragPolarPhi"] = j.DragPolarPhi
		m["dragPolarTheta"] = j.DragPolarTheta
		m["indexPhi"] = j.IndexPhi
		m["indexTheta"] = j.IndexTheta
		m["indexR"] = j.IndexR
		if j.TopTiltVectorPhiIdx != 0 {
			m["topTiltVectorThetaIdx"] = j.TopTiltVectorPhiIdx
		}
	})
}

func Read(root, id string) (JSON, bool) {
	var j JSON
	if !jsonpersist.ReadJSONIfExists(DragPath(root, id), &j) {
		return JSON{}, false
	}
	return j, true
}
