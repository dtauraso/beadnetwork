package edgefile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

type deltaFields struct {
	DeltaPolarR     *float64 `json:"deltaPolarR,omitempty"`
	DeltaPolarPhi   *float64 `json:"deltaPolarPhi,omitempty"`
	DeltaPolarTheta *float64 `json:"deltaPolarTheta,omitempty"`
}

func edgeDragPath(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "drag", "edges", label+".json")
}

func ReadEdgeDelta(root, src, label string) (polar.Polar, bool) {
	var df deltaFields
	if !jsonpersist.ReadJSONIfExists(edgeFilePath(root, src, label), &df) {
		return polar.Polar{}, false
	}
	if df.DeltaPolarR == nil || df.DeltaPolarPhi == nil || df.DeltaPolarTheta == nil {
		return polar.Polar{}, false
	}
	return polar.Polar{R: *df.DeltaPolarR, Phi: *df.DeltaPolarPhi, Theta: *df.DeltaPolarTheta}, true
}

func ReadEdgeDragDelta(root, src, label string) (polar.Polar, bool) {
	var df deltaFields
	if !jsonpersist.ReadJSONIfExists(edgeDragPath(root, src, label), &df) {
		return polar.Polar{}, false
	}
	if df.DeltaPolarR == nil || df.DeltaPolarPhi == nil || df.DeltaPolarTheta == nil {
		return polar.Polar{}, false
	}
	return polar.Polar{R: *df.DeltaPolarR, Phi: *df.DeltaPolarPhi, Theta: *df.DeltaPolarTheta}, true
}

func WriteEdgeDelta(root, src, label string, d polar.Polar) error {
	path := edgeDragPath(root, src, label)
	return jsonpersist.ReadModifyWriteJSON(path, func(m map[string]any) {
		m["deltaPolarR"] = d.R
		m["deltaPolarPhi"] = d.Phi
		m["deltaPolarTheta"] = d.Theta
	})
}
