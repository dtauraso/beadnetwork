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
	IPhi           int     `json:"iPhi"`
	ITheta         int     `json:"iTheta"`
	IR             int     `json:"iR"`

	TopTiltVectorPhiIdx int32 `json:"topTiltVectorThetaIdx,omitempty"`
}

func Write(root, id string, j JSON) error {
	path := DragPath(root, id)
	return jsonpersist.ReadModifyWriteJSON(path, func(m map[string]any) {
		m["dragPolarR"] = j.DragPolarR
		m["dragPolarPhi"] = j.DragPolarPhi
		m["dragPolarTheta"] = j.DragPolarTheta
		m["iPhi"] = j.IPhi
		m["iTheta"] = j.ITheta
		m["iR"] = j.IR
		delete(m, "stepPhi")
		delete(m, "stepTheta")
		delete(m, "stepR")
		delete(m, "quantIPhi")
		delete(m, "quantITheta")
		delete(m, "quantIR")
		delete(m, "deltaPolarR")
		delete(m, "deltaPolarPhi")
		delete(m, "deltaPolarTheta")
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
