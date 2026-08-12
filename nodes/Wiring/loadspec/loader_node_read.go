// loader_node_read.go — reads a single node's own files: its meta.json (+ the position.json
// override and optional data.json), and its edges/ subdir. LoadTree (loader_tree.go) calls
// loadNodeMeta then loadNodeEdges once per node directory, in that order; splitting them out
// here groups "read one node's own files" as its own concern, separate from LoadTree's
// directory-walk orchestration and validateNodeIDs' id validation.
//
// Neither function here constructs a filepath.Join with a literal "nodes" path segment
// (that Join happens once, in LoadTree itself) — nodesDir/nodeDir/edgesDir below are all
// built from an already-resolved directory, so tools/network/persist/check-scene-path-resolution.sh's
// NODE_PATH_OWNERS allowlist does not need a new entry for this file.

package loadspec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
)

// JSONMeta is the shape of nodes/<id>/meta.json.
type JSONMeta struct {
	ID   string   `json:"id"`
	Type string   `json:"type"`
	R    *float64 `json:"r,omitempty"` // optional per-node sphere radius; nil → nodegeom.DefaultNodeR (see nodegeom.NodeR)
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

// loadNodeMeta reads nodeID's meta.json (required — static identity: id/type/r/gate, and
// for a PRE-SPLIT topology also the position fields inline, legacy shape), then its
// position.json (the POST-SPLIT position writer's file, quant_offset_persist.go
// writeQuantOffset — present overrides meta.json's possibly stale/legacy position fields,
// absent leaves sn with whatever meta.json carried), then its optional data.json.
func loadNodeMeta(root, nodesDir, nodeID string) (specNode, error) {
	nodeDir := filepath.Join(nodesDir, nodeID)

	metaPath := filepath.Join(nodeDir, "meta.json")
	metaRaw, err := os.ReadFile(metaPath)
	if err != nil {
		return specNode{}, fmt.Errorf("loadTree: node %q meta: %w", nodeID, err)
	}
	var meta JSONMeta
	if err := json.Unmarshal(metaRaw, &meta); err != nil {
		return specNode{}, fmt.Errorf("loadTree: node %q meta parse: %w", nodeID, err)
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

	var pf positionfile.JSON
	if jsonpersist.ReadJSONIfExists(positionfile.FilePath(root, nodeID), &pf) {
		r, th, ph := pf.ScenePolarR, pf.ScenePolarTheta, pf.ScenePolarPhi
		qt, qp, qr := pf.QuantITheta, pf.QuantIPhi, pf.QuantIR
		st, sp, sr := pf.StepTheta, pf.StepPhi, pf.StepR
		sn.ScenePolarR, sn.ScenePolarTheta, sn.ScenePolarPhi = &r, &th, &ph
		sn.QuantITheta, sn.QuantIPhi, sn.QuantIR = &qt, &qp, &qr
		sn.StepTheta, sn.StepPhi, sn.StepR = &st, &sp, &sr
		vt := pf.TopTiltVectorThetaIdx
		sn.TopTiltVectorThetaIdx = &vt
	}

	dataPath := filepath.Join(nodeDir, "data.json")
	if raw, err := os.ReadFile(dataPath); err == nil {
		var nd NodeData
		if err := json.Unmarshal(raw, &nd); err != nil {
			return specNode{}, fmt.Errorf("loadTree: node %q data parse: %w", nodeID, err)
		}
		sn.Data = &nd
	}

	return sn, nil
}

// loadNodeEdges reads nodeID's edges/ subdir — this node's OUTGOING edges (adjacency
// list). Optional subdir: a node with no outgoing edges simply has no edges/ subdir, which
// is normal, not an error. Order rule: edges appear in outer node-directory-numeric-sorted
// order (established by LoadTree's own nodeDirs iteration), then inner
// sort.Strings(edgeFiles) within each node's edges/. Unlike node ids, edge-file names are
// LABELS, not numbers — a plain lexicographic string sort is correct here and must stay
// that way; do not "fix" this the same way node ids were fixed.
func loadNodeEdges(nodesDir, nodeID string) ([]specEdge, error) {
	nodeDir := filepath.Join(nodesDir, nodeID)
	edgesDir := filepath.Join(nodeDir, "edges")
	edgeFiles, err := readDirNames(edgesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("loadTree: list node %q edges dir %s: %w", nodeID, edgesDir, err)
	}
	sort.Strings(edgeFiles)

	var edges []specEdge
	for _, fname := range edgeFiles {
		if !strings.HasSuffix(fname, ".json") {
			continue
		}
		fpath := filepath.Join(edgesDir, fname)
		raw, err := os.ReadFile(fpath)
		if err != nil {
			return nil, fmt.Errorf("loadTree: read edge file %s: %w", fpath, err)
		}
		var e specEdge
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("loadTree: parse edge file %s: %w", fpath, err)
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
		edges = append(edges, e)
	}
	return edges, nil
}
