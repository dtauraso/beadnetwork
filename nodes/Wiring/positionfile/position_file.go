package positionfile

import "path/filepath"

func FilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "position.json")
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

	TopTiltVectorThetaIdx int32 `json:"topTiltVectorThetaIdx,omitempty"`
}
