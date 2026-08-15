package edgefile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

type deltaFields struct {
	DeltaIndexR     *int `json:"deltaIndexR,omitempty"`
	DeltaIndexPhi   *int `json:"deltaIndexPhi,omitempty"`
	DeltaIndexTheta *int `json:"deltaIndexTheta,omitempty"`
}

func edgeDragPath(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "drag", "edges", label+".json")
}

func ReadEdgeDelta(root, src, label string, sc polarindex.SceneConstants) (polar.Polar, bool) {
	var df deltaFields
	if !jsonpersist.ReadJSONIfExists(edgeFilePath(root, src, label), &df) {
		return polar.Polar{}, false
	}
	if df.DeltaIndexR == nil || df.DeltaIndexPhi == nil || df.DeltaIndexTheta == nil {
		return polar.Polar{}, false
	}
	idx := polarindex.Index{Phi: *df.DeltaIndexPhi, Theta: *df.DeltaIndexTheta, R: *df.DeltaIndexR}
	return polarindex.ToPolar(idx, sc), true
}

type dragIndexFields struct {
	IndexPhi   *int `json:"indexPhi,omitempty"`
	IndexTheta *int `json:"indexTheta,omitempty"`
	IndexR     *int `json:"indexR,omitempty"`
}

func ReadEdgeDragIndex(root, src, label string) (polarindex.Index, bool) {
	var df dragIndexFields
	if !jsonpersist.ReadJSONIfExists(edgeDragPath(root, src, label), &df) {
		return polarindex.Index{}, false
	}
	if df.IndexPhi == nil || df.IndexTheta == nil || df.IndexR == nil {
		return polarindex.Index{}, false
	}
	return polarindex.Index{Phi: *df.IndexPhi, Theta: *df.IndexTheta, R: *df.IndexR}, true
}

func WriteEdgeDrag(root, src, label string, idx polarindex.Index) error {
	path := edgeDragPath(root, src, label)
	return jsonpersist.ReadModifyWriteJSON(path, func(m map[string]any) {
		m["indexPhi"] = idx.Phi
		m["indexTheta"] = idx.Theta
		m["indexR"] = idx.R
		delete(m, "dragPolarR")
		delete(m, "dragPolarPhi")
		delete(m, "dragPolarTheta")
		delete(m, "deltaPolarR")
		delete(m, "deltaPolarPhi")
		delete(m, "deltaPolarTheta")
	})
}
