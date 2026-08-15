package positionfile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/gitskip"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

func FilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "meta.json")
}

type JSON struct {
	ScenePolarR     float64 `json:"scenePolarR"`
	ScenePolarPhi   float64 `json:"scenePolarPhi"`
	ScenePolarTheta float64 `json:"scenePolarTheta"`
	QuantIPhi       int     `json:"quantIPhi"`
	QuantITheta     int     `json:"quantITheta"`
	QuantIR         int     `json:"quantIR"`
	StepPhi         float64 `json:"stepPhi"`
	StepTheta       float64 `json:"stepTheta"`
	StepR           float64 `json:"stepR"`

	TopTiltVectorPhiIdx int32 `json:"topTiltVectorThetaIdx,omitempty"`
}

func Write(root, id string, j JSON) error {
	path := FilePath(root, id)
	if err := jsonpersist.ReadModifyWriteJSON(path, func(m map[string]any) {
		m["scenePolarR"] = j.ScenePolarR
		m["scenePolarPhi"] = j.ScenePolarPhi
		m["scenePolarTheta"] = j.ScenePolarTheta
		m["quantIPhi"] = j.QuantIPhi
		m["quantITheta"] = j.QuantITheta
		m["quantIR"] = j.QuantIR
		m["stepPhi"] = j.StepPhi
		m["stepTheta"] = j.StepTheta
		m["stepR"] = j.StepR
		if j.TopTiltVectorPhiIdx != 0 {
			m["topTiltVectorThetaIdx"] = j.TopTiltVectorPhiIdx
		}
	}); err != nil {
		return err
	}
	gitskip.Mark(path)
	return nil
}
