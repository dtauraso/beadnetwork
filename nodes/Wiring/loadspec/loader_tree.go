package loadspec

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func LoadTree(root string) (TopoSpec, error) {
	var spec TopoSpec

	nodesDir := filepath.Join(root, "nodes")
	nodeDirs, err := readDirNames(nodesDir)
	if err != nil {
		return spec, fmt.Errorf("loadTree: list nodes dir %s: %w", nodesDir, err)
	}

	rowCount, err := validateNodeIDs(nodeDirs)
	if err != nil {
		return spec, err
	}
	spec.RowCount = rowCount

	for _, nodeID := range nodeDirs {
		sn, err := loadNodeMeta(root, nodesDir, nodeID)
		if err != nil {
			return spec, err
		}
		spec.Nodes = append(spec.Nodes, sn)

		edges, err := loadNodeEdges(root, nodesDir, nodeID)
		if err != nil {
			return spec, err
		}
		spec.Edges = append(spec.Edges, edges...)
	}

	ResolveEdgeDeltas(&spec)
	PlaceFromDeltas(&spec)

	return spec, nil
}

func validateNodeIDs(nodeDirs []string) (int, error) {
	rowCount := 0
	seenIDs := make(map[int]string, len(nodeDirs))
	for _, name := range nodeDirs {
		n, err := strconv.Atoi(name)
		if err != nil {
			return 0, fmt.Errorf("loadTree: node directory %q is not a numeric id: %w", name, err)
		}
		if n < 1 {
			return 0, fmt.Errorf("loadTree: node directory %q has id %d, but node ids are 1-based (must be >= 1)", name, n)
		}
		if prev, dup := seenIDs[n]; dup {
			return 0, fmt.Errorf("loadTree: node directories %q and %q both parse to id %d — duplicate node id", prev, name, n)
		}
		seenIDs[n] = name
		if n > rowCount {
			rowCount = n
		}
	}
	return rowCount, nil
}

func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}
