package edgefile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

func DeltaFilePath(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "edges", label+".geom.json")
}

type deltaFile struct {
	DeltaPolarR     float64 `json:"deltaPolarR"`
	DeltaPolarPhi   float64 `json:"deltaPolarPhi"`
	DeltaPolarTheta float64 `json:"deltaPolarTheta"`
}

func ReadEdgeDelta(root, src, label string) (polar.Polar, bool) {
	var df deltaFile
	if !jsonpersist.ReadJSONIfExists(DeltaFilePath(root, src, label), &df) {
		return polar.Polar{}, false
	}
	return polar.Polar{R: df.DeltaPolarR, Phi: df.DeltaPolarPhi, Theta: df.DeltaPolarTheta}, true
}

func WriteEdgeDelta(root, src, label string, d polar.Polar) error {
	return jsonpersist.WriteJSONAtomic(DeltaFilePath(root, src, label),
		deltaFile{DeltaPolarR: d.R, DeltaPolarPhi: d.Phi, DeltaPolarTheta: d.Theta})
}
