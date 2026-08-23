package structuraledit

import (
	"github.com/dtauraso/wirefold/Categories/Node"
	"path/filepath"
)

func nodesPath(root string) string { return filepath.Join(root, "counts", "nodes.bin") }

func edgesPath(root string) string { return filepath.Join(root, "counts", "edges.bin") }

func writeCounts(root string, nodes, edges int) error {
	if err := Node.WriteAtomic(nodesPath(root), nodes); err != nil {
		return err
	}
	return Node.WriteAtomic(edgesPath(root), edges)
}
