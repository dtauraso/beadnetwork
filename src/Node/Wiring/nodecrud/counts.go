package nodecrud

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/Node/Wiring/jsonpersist"
)

func nodesPath(root string) string { return filepath.Join(root, "counts", "nodes.json") }

func edgesPath(root string) string { return filepath.Join(root, "counts", "edges.json") }

func writeCounts(root string, nodes, edges int) error {
	if err := jsonpersist.WriteJSONAtomic(nodesPath(root), nodes); err != nil {
		return err
	}
	return jsonpersist.WriteJSONAtomic(edgesPath(root), edges)
}
