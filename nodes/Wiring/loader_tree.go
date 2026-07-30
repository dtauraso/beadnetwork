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
	// LocalPolars — see specNode's field doc (loader.go).
	LocalPolars []specLocalPolar `json:"localPolars,omitempty"`
	// LocalPoleTheta/LocalPolePhi — see specNode's field doc (loader.go).
	LocalPoleTheta *float64 `json:"localPoleTheta,omitempty"`
	LocalPolePhi   *float64 `json:"localPolePhi,omitempty"`
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
	sort.Strings(nodeDirs)

	for _, nodeID := range nodeDirs {
		nodeDir := filepath.Join(nodesDir, nodeID)

		// meta.json — required. Still owns static node identity (id/type/r/gate) and, for a
		// PRE-SPLIT topology, also the position/local-polars fields inline (legacy shape).
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
			LocalPolars:     meta.LocalPolars,
			LocalPoleTheta:  meta.LocalPoleTheta,
			LocalPolePhi:    meta.LocalPolePhi,
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
		}

		// local-polars.json — the POST-SPLIT local-polars writer's file (quant_offset_persist.go
		// WriteLocalPolars). Present → overrides meta.json's legacy localPolars/pole fields.
		var lpf localPolarsFileJSON
		readJSONBestEffort(localPolarsFilePath(root, nodeID), &lpf)
		if lpf.LocalPolars != nil {
			lps := make([]specLocalPolar, 0, len(lpf.LocalPolars))
			for _, lp := range lpf.LocalPolars {
				lps = append(lps, specLocalPolar(lp))
			}
			sn.LocalPolars = lps
			pt, pp := lpf.LocalPoleTheta, lpf.LocalPolePhi
			sn.LocalPoleTheta, sn.LocalPolePhi = &pt, &pp
		}

		// cascade-edges.json — this node's STORED cascade-neighbor id list (specNode.
		// CascadeEdges doc comment). Missing file → empty list, not an error (readJSONBestEffort's
		// standard missing-file default).
		var cef cascadeEdgesFileJSON
		readJSONBestEffort(cascadeEdgesFilePath(root, nodeID), &cef)
		sn.CascadeEdges = cef.CascadeEdges
		sn.CascadeKinds = cef.CascadeKinds

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
		// error. Order rule: edges appear in outer
		// node-directory-sorted order (already established by the sort.Strings(nodeDirs)
		// above), then inner sort.Strings(edgeFiles) within each node's edges/ — i.e.
		// deterministic by (source id, label) lexical order, not by label alone as the
		// flat single-dir layout gave before this change.
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
