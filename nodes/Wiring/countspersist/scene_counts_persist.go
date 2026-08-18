package countspersist

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

func NodesPath(root string) string { return filepath.Join(root, "counts", "nodes.json") }

func EdgesPath(root string) string { return filepath.Join(root, "counts", "edges.json") }

func WriteCounts(root string, nodes, edges int) error {
	if err := jsonpersist.WriteJSONAtomic(NodesPath(root), nodes); err != nil {
		return err
	}
	return jsonpersist.WriteJSONAtomic(EdgesPath(root), edges)
}
