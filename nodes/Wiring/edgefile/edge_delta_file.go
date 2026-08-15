package edgefile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

func edgeDragPath(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "drag", "edges", label+".json")
}

type dragIndexFile struct {
	IndexPhi   int `json:"indexPhi"`
	IndexTheta int `json:"indexTheta"`
	IndexR     int `json:"indexR"`
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
	return jsonpersist.WriteJSONAtomic(path, dragIndexFile{
		IndexPhi:   idx.Phi,
		IndexTheta: idx.Theta,
		IndexR:     idx.R,
	})
}
