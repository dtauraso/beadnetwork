package positionfile

import "path/filepath"

func FilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "position.json")
}

type JSON struct {
	ScenePolarR     float64 `json:"scenePolarR"`
	ScenePolarTheta float64 `json:"scenePolarTheta"`
	ScenePolarPhi   float64 `json:"scenePolarPhi"`
	QuantITheta     int     `json:"quantITheta"`
	QuantIPhi       int     `json:"quantIPhi"`
	QuantIR         int     `json:"quantIR"`
	StepTheta       float64 `json:"stepTheta"`
	StepPhi         float64 `json:"stepPhi"`
	StepR           float64 `json:"stepR"`

	TopTiltVectorThetaIdx int32 `json:"topTiltVectorThetaIdx,omitempty"`
}
