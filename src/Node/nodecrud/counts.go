package nodecrud

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

func nodesPath(root string) string { return filepath.Join(root, "counts", "nodes.bin") }

func edgesPath(root string) string { return filepath.Join(root, "counts", "edges.bin") }

func writeCounts(root string, nodes, edges int) error {
	if err := valuefile.WriteAtomic(nodesPath(root), nodes); err != nil {
		return err
	}
	return valuefile.WriteAtomic(edgesPath(root), edges)
}
