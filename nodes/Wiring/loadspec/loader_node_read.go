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

type JSONMeta struct {
	ID   string   `json:"id"`
	Type string   `json:"type"`
	R    *float64 `json:"r,omitempty"`

	ScenePolarR     *float64 `json:"scenePolarR,omitempty"`
	ScenePolarTheta *float64 `json:"scenePolarTheta,omitempty"`
	ScenePolarPhi   *float64 `json:"scenePolarPhi,omitempty"`

	QuantITheta *int `json:"quantITheta,omitempty"`
	QuantIPhi   *int `json:"quantIPhi,omitempty"`
	QuantIR     *int `json:"quantIR,omitempty"`

	StepTheta *float64 `json:"stepTheta,omitempty"`
	StepPhi   *float64 `json:"stepPhi,omitempty"`
	StepR     *float64 `json:"stepR,omitempty"`

	Gate bool `json:"gate,omitempty"`
}

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
