package nodecrud

import (
	"path/filepath"
)

func nodesPath(root string) string { return filepath.Join(root, "counts", "nodes.bin") }

func edgesPath(root string) string { return filepath.Join(root, "counts", "edges.bin") }

func writeCounts(root string, nodes, edges int) error {
	if err := WriteAtomic(nodesPath(root), nodes); err != nil {
		return err
	}
	return WriteAtomic(edgesPath(root), edges)
}
