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
// There is no nodes/<id>/inputs/ or outputs/ any more (docs/channels-not-ports.md): a
// port is a load-time channel-binding ROLE resolved from the kind's registry
// (PortSpec/a.In()/a.Out()), never a placed entity with its own geometry file.
//
// It returns a topoSpec in the same shape parseSpec/LoadTopology consume regardless of
// how the tree was read.

package Wiring

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// jsonMeta is the shape of nodes/<id>/meta.json.
type jsonMeta struct {
	ID   string   `json:"id"`
	Type string   `json:"type"`
	R    *float64 `json:"r,omitempty"` // optional per-node sphere radius; nil → defaultNodeR (see nodeR)
	// Scene polar (polar-frame-rewrite.md) — the node's position as (r,θ,φ) about the scene
	// sphere center. This is the authoritative stored position. These MUST be carried through
	// to specNode below: dropping them collapses every node to the origin (the blob bug).
	ScenePolarR     *float64 `json:"scenePolarR,omitempty"`
	ScenePolarTheta *float64 `json:"scenePolarTheta,omitempty"`
	ScenePolarPhi   *float64 `json:"scenePolarPhi,omitempty"`
	// Quantized polar offset (quantized_layout.go PHASE 3) — see specNode's field doc.
	QuantITheta *int `json:"quantITheta,omitempty"`
	QuantIPhi   *int `json:"quantIPhi,omitempty"`
	QuantIR     *int `json:"quantIR,omitempty"`
	// Per-node step constants — see specNode's field doc (loader.go).
	StepTheta *float64 `json:"stepTheta,omitempty"`
	StepPhi   *float64 `json:"stepPhi,omitempty"`
	StepR     *float64 `json:"stepR,omitempty"`
	// Gate — see specNode's field doc (loader.go). UNCONSUMED since the
	// rule/gate/anchor cascade was deleted (2026-07-18): round-tripped to/from
	// meta.json only, no code path reads it for behavior.
	Gate bool `json:"gate,omitempty"`
}

// loadTree reads the directory-tree topology rooted at root and assembles a
// topoSpec.  All subdirectory entries are sorted so the result is deterministic.
func loadTree(root string) (topoSpec, error) {
	var spec topoSpec

	// ── nodes ────────────────────────────────────────────────────────────────
	nodesDir := filepath.Join(root, "nodes")
	nodeDirs, err := readDirNames(nodesDir)
	if err != nil {
		return spec, fmt.Errorf("loadTree: list nodes dir %s: %w", nodesDir, err)
	}
	// ROW ID = NODE ID - 1 (declared by the directory name, never derived by sorting). Node
	// ids ARE numbers — they only appear as strings because they are directory names — and
	// node identity IS the buffer row index (no id sidecar): a row is decided by parsing the
	// directory name, not by where it falls after a sort. A directory name that isn't a
	// number, an id below 1 (ids are 1-based), or a duplicate parsed id is a load error, loud
	// and naming the offending directory — never a silent fallback. The row space itself
	// (spec.RowCount) is sized by the LARGEST id found, not by the node count: a deleted node
	// leaves its row empty rather than collapsing later rows upward — that collapse is
	// precisely the silent renaming this model removes. There is no ordering left to assert:
	// loop order below only affects the order edges are appended to spec.Edges, which carries
	// no row semantics of its own.
	seenIDs := make(map[int]string, len(nodeDirs))
	for _, name := range nodeDirs {
		n, err := strconv.Atoi(name)
		if err != nil {
			return spec, fmt.Errorf("loadTree: node directory %q is not a numeric id: %w", name, err)
		}
		if n < 1 {
			return spec, fmt.Errorf("loadTree: node directory %q has id %d, but node ids are 1-based (must be >= 1)", name, n)
		}
		if prev, dup := seenIDs[n]; dup {
			return spec, fmt.Errorf("loadTree: node directories %q and %q both parse to id %d — duplicate node id", prev, name, n)
		}
		seenIDs[n] = name
		if n > spec.RowCount {
			spec.RowCount = n
		}
	}

	for _, nodeID := range nodeDirs {
		nodeDir := filepath.Join(nodesDir, nodeID)

		// meta.json — required. Still owns static node identity (id/type/r/gate) and, for a
		// PRE-SPLIT topology, also the position fields inline (legacy shape).
		metaPath := filepath.Join(nodeDir, "meta.json")
		metaRaw, err := os.ReadFile(metaPath)
		if err != nil {
			return spec, fmt.Errorf("loadTree: node %q meta: %w", nodeID, err)
		}
		var meta jsonMeta
		if err := json.Unmarshal(metaRaw, &meta); err != nil {
			return spec, fmt.Errorf("loadTree: node %q meta parse: %w", nodeID, err)
		}

		sn := specNode{
			ID:              meta.ID,
			Type:            meta.Type,
			R:               meta.R,
			ScenePolarR:     meta.ScenePolarR,
			ScenePolarTheta: meta.ScenePolarTheta,
			ScenePolarPhi:   meta.ScenePolarPhi,
			QuantITheta:     meta.QuantITheta,
			QuantIPhi:       meta.QuantIPhi,
			QuantIR:         meta.QuantIR,
			StepTheta:       meta.StepTheta,
			StepPhi:         meta.StepPhi,
			StepR:           meta.StepR,
			Gate:            meta.Gate,
		}

		// position.json — the POST-SPLIT position writer's file (quant_offset_persist.go
		// writeQuantOffset). Present → overrides meta.json's (possibly stale/legacy) position
		// fields. Absent → sn keeps whatever meta.json carried above (old-format topology).
		var pf positionFileJSON
		if readJSONIfExists(positionFilePath(root, nodeID), &pf) {
			r, th, ph := pf.ScenePolarR, pf.ScenePolarTheta, pf.ScenePolarPhi
			qt, qp, qr := pf.QuantITheta, pf.QuantIPhi, pf.QuantIR
			st, sp, sr := pf.StepTheta, pf.StepPhi, pf.StepR
			sn.ScenePolarR, sn.ScenePolarTheta, sn.ScenePolarPhi = &r, &th, &ph
			sn.QuantITheta, sn.QuantIPhi, sn.QuantIR = &qt, &qp, &qr
			sn.StepTheta, sn.StepPhi, sn.StepR = &st, &sp, &sr
			vt, vp := pf.VectorThetaIdx, pf.VectorPhiIdx
			sn.VectorThetaIdx, sn.VectorPhiIdx = &vt, &vp
		}

		// data.json — optional
		dataPath := filepath.Join(nodeDir, "data.json")
		if raw, err := os.ReadFile(dataPath); err == nil {
			var nd NodeData
			if err := json.Unmarshal(raw, &nd); err != nil {
				return spec, fmt.Errorf("loadTree: node %q data parse: %w", nodeID, err)
			}
			sn.Data = &nd
		}

		spec.Nodes = append(spec.Nodes, sn)

		// edges/ — this node's OUTGOING edges (adjacency list). Optional subdir: a node
		// with no outgoing edges simply has no edges/ subdir, which is normal, not an
		// error. Order rule: edges appear in outer node-directory-numeric-sorted order
		// (already established above), then inner sort.Strings(edgeFiles) within each
		// node's edges/. Unlike node ids, edge-file names are LABELS, not numbers — a
		// plain lexicographic string sort is correct here and must stay that way; do not
		// "fix" this the same way node ids were fixed.
		edgesDir := filepath.Join(nodeDir, "edges")
		edgeFiles, err := readDirNames(edgesDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return spec, fmt.Errorf("loadTree: list node %q edges dir %s: %w", nodeID, edgesDir, err)
		}
		sort.Strings(edgeFiles)

		for _, fname := range edgeFiles {
			if !strings.HasSuffix(fname, ".json") {
				continue
			}
			fpath := filepath.Join(edgesDir, fname)
			raw, err := os.ReadFile(fpath)
			if err != nil {
				return spec, fmt.Errorf("loadTree: read edge file %s: %w", fpath, err)
			}
			var e specEdge
			if err := json.Unmarshal(raw, &e); err != nil {
				return spec, fmt.Errorf("loadTree: parse edge file %s: %w", fpath, err)
			}
			// The source is the directory the file sits in, not a field left inside the
			// file. A stale "source" key that disagrees with nodeID means the file was
			// moved (or hand-edited) without updating its content — trust the directory,
			// which is the addressing scheme, and fail loudly rather than silently
			// accepting a value that could drift from where the file actually lives.
			if e.Source != "" && e.Source != nodeID {
				panic(fmt.Sprintf(
					"loadTree: edge file %s has stale source %q that disagrees with its "+
						"containing node directory %q — the adjacency layout (topology/nodes/<id>/edges/) "+
						"derives an edge's source from the directory it is stored under, not from a "+
						"\"source\" key in the file; whatever wrote/moved this file should have dropped "+
						"the redundant source key or kept it in sync with the directory",
					fpath, e.Source, nodeID))
			}
			e.Source = nodeID
			spec.Edges = append(spec.Edges, e)
		}
	}

	return spec, nil
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
