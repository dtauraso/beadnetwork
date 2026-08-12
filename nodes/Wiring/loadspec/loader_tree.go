// loader_tree.go — directory-tree topology reader.
//
// loadTree reads a topology laid out as a directory tree:
//
//	<root>/nodes/<id>/meta.json         — id, type
//	<root>/nodes/<id>/data.json         — NodeData (optional)
//	<root>/nodes/<id>/edges/<label>.json  — specEdge, OUTGOING only (adjacency list: the
//	                                        source is the directory the file sits in, not a
//	                                        field in the file — see specEdge's doc comment)
//
// There is no nodes/<id>/inputs/ or outputs/ any more (docs/bead-model/channels-not-ports.md): a
// port is a load-time channel-binding ROLE resolved from the kind's registry
// (PortSpec/a.In()/a.Out()), never a placed entity with its own geometry file.
//
// It returns a TopoSpec in the same shape ParseSpec/LoadTopology consume regardless of
// how the tree was read.
//
// This file holds LoadTree itself (the orchestrator), id validation (validateNodeIDs), and
// the generic directory-listing helper (readDirNames). Reading a single node's own files —
// its meta.json/position.json/data.json and its edges/ subdir — is loader_node_read.go's
// concern (loadNodeMeta/loadNodeEdges); LoadTree calls both per node directory, in order.

package loadspec

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// LoadTree reads the directory-tree topology rooted at root and assembles a
// TopoSpec.  All subdirectory entries are sorted so the result is deterministic.
func LoadTree(root string) (TopoSpec, error) {
	var spec TopoSpec

	// ── nodes ────────────────────────────────────────────────────────────────
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

		edges, err := loadNodeEdges(nodesDir, nodeID)
		if err != nil {
			return spec, err
		}
		spec.Edges = append(spec.Edges, edges...)
	}

	return spec, nil
}

// validateNodeIDs parses every node directory name to its numeric id and returns the row
// space size (spec.RowCount).
//
// ROW ID = NODE ID - 1 (declared by the directory name, never derived by sorting). Node
// ids ARE numbers — they only appear as strings because they are directory names — and
// node identity IS the buffer row index (no id sidecar): a row is decided by parsing the
// directory name, not by where it falls after a sort. A directory name that isn't a
// number, an id below 1 (ids are 1-based), or a duplicate parsed id is a load error, loud
// and naming the offending directory — never a silent fallback. The row space itself
// (spec.RowCount) is sized by the LARGEST id found, not by the node count: a deleted node
// leaves its row empty rather than collapsing later rows upward — that collapse is
// precisely the silent renaming this model removes. There is no ordering left to assert:
// loop order in LoadTree only affects the order edges are appended to spec.Edges, which
// carries no row semantics of its own.
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

// readDirNames returns the names (not full paths) of all entries in dir.
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
