package positionfile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

func DragPath(root, id string) string {
	return filepath.Join(root, "nodes", id, "drag", "position.json")
}

type JSON struct {
	DeltaPolarR     float64 `json:"deltaPolarR"`
	DeltaPolarPhi   float64 `json:"deltaPolarPhi"`
	DeltaPolarTheta float64 `json:"deltaPolarTheta"`
	QuantIPhi       int     `json:"quantIPhi"`
	QuantITheta     int     `json:"quantITheta"`
	QuantIR         int     `json:"quantIR"`
	StepPhi         float64 `json:"stepPhi"`
	StepTheta       float64 `json:"stepTheta"`
	StepR           float64 `json:"stepR"`

	TopTiltVectorPhiIdx int32 `json:"topTiltVectorThetaIdx,omitempty"`
}

func Write(root, id string, j JSON) error {
	path := DragPath(root, id)
	return jsonpersist.ReadModifyWriteJSON(path, func(m map[string]any) {
		m["deltaPolarR"] = j.DeltaPolarR
		m["deltaPolarPhi"] = j.DeltaPolarPhi
		m["deltaPolarTheta"] = j.DeltaPolarTheta
		m["quantIPhi"] = j.QuantIPhi
		m["quantITheta"] = j.QuantITheta
		m["quantIR"] = j.QuantIR
		m["stepPhi"] = j.StepPhi
		m["stepTheta"] = j.StepTheta
		m["stepR"] = j.StepR
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
