package structuraledit

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodefile"
	"path/filepath"
)

func nodesPath(root string) string { return filepath.Join(root, "counts", "nodes.bin") }

func edgesPath(root string) string { return filepath.Join(root, "counts", "edges.bin") }

func writeCounts(root string, nodes, edges int) error {
	if err := nodefile.WriteAtomic(nodesPath(root), nodes); err != nil {
		return err
	}
	return nodefile.WriteAtomic(edgesPath(root), edges)
}
