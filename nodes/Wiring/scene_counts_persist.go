// scene_counts_persist.go — the ONLY writer of counts.json.
//
// counts.json is its own ownership category, which is why it gets its own file the way each
// view/ file does. It is neither per-node nor per-edge nor view state: it is the TREE'S
// SHAPE, and it exists solely because the extension host must size its stdio array — one
// dedicated pipe per emitting goroutine — BEFORE Go is running to be asked
// (.claude/rules/persistence-ownership.md, "Counts are stored, never re-derived").
//
// Its rule is SINGLE-WRITER: the operation that creates, deletes or renumbers a node is the
// operation that updates it, and nothing else touches it. That operation is
// scene_structure.go's CreateNode/DeleteNode, and until they existed nothing in Go wrote
// this file at all — it was maintained by hand alongside the tree.
package Wiring

import "path/filepath"

type countsFile struct {
	Nodes int `json:"nodes"`
	Edges int `json:"edges"`
}

// WriteCounts rewrites counts.json.
//
// `nodes` is the LARGEST node id (the row count under ROW ID = NODE ID - 1), never the
// number of directories: a gap left by a deleted node still needs its row, and its dedicated
// fd, allocated. `edges` is a plain count — an edge has no id space to leave gaps in.
func WriteCounts(root string, nodes, edges int) error {
	return writeJSONAtomic(filepath.Join(root, "counts.json"), countsFile{Nodes: nodes, Edges: edges})
}
