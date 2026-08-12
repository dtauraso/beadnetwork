package countspersist

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

type countsFile struct {
	Nodes int `json:"nodes"`
	Edges int `json:"edges"`
}

func WriteCounts(root string, nodes, edges int) error {
	return jsonpersist.WriteJSONAtomic(filepath.Join(root, "counts.json"), countsFile{Nodes: nodes, Edges: edges})
}
