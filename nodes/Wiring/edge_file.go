// edge_file.go — an edge's on-disk shape and the two operations that touch it directly:
// WriteEdgeFile (create) and RemoveEdgesTo (delete-by-target). Pure persistence, no
// edgeMover state here — see edge_mover.go for the runtime actor that owns THIS edge's
// live geometry and beads.
//
// An edge is stored under its SOURCE node (<root>/nodes/<src>/edges/<label>.json, outgoing
// only) and carries no `source` key: that is the directory it sits in, and a second copy
// could drift. These helpers live here because a per-edge file may only be written by
// edge_file.go (check-persist-write-ownership) — the palette's create and delete write
// through them rather than constructing an edge path of their own.

package Wiring

import (
	"os"
	"path/filepath"
)

// edgeFile mirrors the on-disk shape exactly (see any topology/nodes/*/edges/*.json).
type edgeFile struct {
	SourceHandle string `json:"sourceHandle"`
	Target       string `json:"target"`
	TargetHandle string `json:"targetHandle"`
	Kind         string `json:"kind"`
	Label        string `json:"label"`
}

// edgeFilePath is <root>/nodes/<src>/edges/<label>.json.
func edgeFilePath(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "edges", label+".json")
}

// edgesDirPath is <root>/nodes/<id>/edges — one node's OUTGOING edges.
func edgesDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id, "edges")
}

// WriteEdgeFile writes the edge src->target under src. The label encodes both endpoints
// ("1To2"), the convention the whole tree already uses.
//
// THE HANDLES ARE PASSED IN, not assumed. They used to be hardcoded "Out" and "In", which is
// true of most kinds and false of some: NormalSum's inputs are NormalA/NormalB, so an edge
// written that way named a port that does not exist. Nothing rejected it at the drop —
// loading is what rejects it, and by then the process is the RESPAWNED one, which exits
// during load. On screen that is the editor freezing: Go owns the camera, so a Go that
// exited is a scene with no zoom, pan or rotate and nothing saying why.
func WriteEdgeFile(root, src, srcPort, target, targetPort string) error {
	label := src + "To" + target
	return writeJSONAtomic(edgeFilePath(root, src, label), edgeFile{
		SourceHandle: srcPort, Target: target, TargetHandle: targetPort, Kind: "chain", Label: label,
	})
}

// RemoveEdgesTo deletes every edge in the tree whose TARGET is id — the in-edges, which live
// under their own sources rather than under the node they point at. That is a walk, and it
// is the cost of not duplicating an edge under both endpoints; the loader makes the same
// pass to build its inbound map.
func RemoveEdgesTo(root, id string, nodeIDs []string) error {
	for _, n := range nodeIDs {
		entries, err := os.ReadDir(edgesDirPath(root, n))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() { // path-resolution-ok: skipping a stray directory, not resolving a scene path
				continue
			}
			path := filepath.Join(edgesDirPath(root, n), e.Name())
			var ef edgeFile
			readJSONBestEffort(path, &ef)
			if ef.Target != id {
				continue
			}
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}
