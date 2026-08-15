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

func ReadEdgeDragIndex(root, src, label string) (polarindex.Offset, bool) {
	var df dragIndexFields
	if !jsonpersist.ReadJSONIfExists(edgeDragPath(root, src, label), &df) {
		return polarindex.Offset{}, false
	}
	if df.IndexPhi == nil || df.IndexTheta == nil || df.IndexR == nil {
		return polarindex.Offset{}, false
	}
	return polarindex.Offset{Phi: *df.IndexPhi, Theta: *df.IndexTheta, R: *df.IndexR}, true
}

func WriteEdgeDrag(root, src, label string, off polarindex.Offset) error {
	path := edgeDragPath(root, src, label)
	return jsonpersist.WriteJSONAtomic(path, dragIndexFile{
		IndexPhi:   off.Phi,
		IndexTheta: off.Theta,
		IndexR:     off.R,
	})
}
