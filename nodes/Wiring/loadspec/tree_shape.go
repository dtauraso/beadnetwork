package loadspec

// tree_shape.go — the tree's own shape, for the operations that CHANGE it.
//
// Reading how many nodes and edges a tree has belongs here, beside the loader that reads the
// tree itself (loader_tree.go). scene_structure.go (the palette's create/delete) decides
// what to change; these answer what is there.
//
// counts.json is WRITTEN by scene_counts_persist.go, not here — it is its own owner, the
// way each view/ file has its own persister. This file only reads the tree's shape.

import (
	"path/filepath"
	"sort"
	"strconv"
)

// NodeIDsInTree lists the parsed node ids, ascending. A directory name that does not parse
// as a number is skipped rather than reported — loadTree already fails loudly on one, and
// this is not the place that decides a tree is malformed.
func NodeIDsInTree(root string) []int {
	names, err := readDirNames(filepath.Join(root, "nodes"))
	if err != nil {
		return nil
	}
	out := make([]int, 0, len(names))
	for _, n := range names {
		if v, err := strconv.Atoi(n); err == nil {
			out = append(out, v)
		}
	}
	sort.Ints(out)
	return out
}

// NodeIDStringsInTree is NodeIDsInTree as the strings the rest of the codebase keys by (a
// node id is a string everywhere downstream; only the row derivation parses it).
func NodeIDStringsInTree(root string) []string {
	ids := NodeIDsInTree(root)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, strconv.Itoa(id))
	}
	return out
}

// LargestNodeID is counts.json's "nodes": the LARGEST id, which is the ROW COUNT under
// ROW ID = NODE ID - 1 — not how many directories exist. A gap in the id space still needs
// its row, and its dedicated fd, allocated.
func LargestNodeID(root string) int {
	best := 0
	for _, id := range NodeIDsInTree(root) {
		if id > best {
			best = id
		}
	}
	return best
}

// CountEdgeFiles is counts.json's "edges": a plain count, since an edge has no id space to
// leave gaps in.
func CountEdgeFiles(root string) int {
	total := 0
	for _, id := range NodeIDStringsInTree(root) {
		names, err := readDirNames(filepath.Join(root, "nodes", id, "edges"))
		if err != nil {
			continue
		}
		total += len(names)
	}
	return total
}
