package loadspec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
)

type JSONMeta struct {
	ID   string   `json:"id"`
	Type string   `json:"type"`
	R    *float64 `json:"r,omitempty"`

	ScenePolarR     *float64 `json:"scenePolarR,omitempty"`
	ScenePolarPhi   *float64 `json:"scenePolarPhi,omitempty"`
	ScenePolarTheta *float64 `json:"scenePolarTheta,omitempty"`

	QuantIPhi   *int `json:"quantIPhi,omitempty"`
	QuantITheta *int `json:"quantITheta,omitempty"`
	QuantIR     *int `json:"quantIR,omitempty"`

	StepPhi   *float64 `json:"stepPhi,omitempty"`
	StepTheta *float64 `json:"stepTheta,omitempty"`
	StepR     *float64 `json:"stepR,omitempty"`

	Gate bool `json:"gate,omitempty"`

	Orbit *polar.OrbitRule `json:"orbit,omitempty"`
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
		ScenePolarPhi:   meta.ScenePolarPhi,
		ScenePolarTheta: meta.ScenePolarTheta,
		QuantIPhi:       meta.QuantIPhi,
		QuantITheta:     meta.QuantITheta,
		QuantIR:         meta.QuantIR,
		StepPhi:         meta.StepPhi,
		StepTheta:       meta.StepTheta,
		StepR:           meta.StepR,
		Gate:            meta.Gate,
		Orbit:           meta.Orbit,
	}

	var pf positionfile.JSON
	if jsonpersist.ReadJSONIfExists(positionfile.FilePath(root, nodeID), &pf) {
		r, th, ph := pf.ScenePolarR, pf.ScenePolarPhi, pf.ScenePolarTheta
		qt, qp, qr := pf.QuantIPhi, pf.QuantITheta, pf.QuantIR
		st, sp, sr := pf.StepPhi, pf.StepTheta, pf.StepR
		sn.ScenePolarR, sn.ScenePolarPhi, sn.ScenePolarTheta = &r, &th, &ph
		sn.QuantIPhi, sn.QuantITheta, sn.QuantIR = &qt, &qp, &qr
		sn.StepPhi, sn.StepTheta, sn.StepR = &st, &sp, &sr
		vt := pf.TopTiltVectorPhiIdx
		sn.TopTiltVectorPhiIdx = &vt
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

func loadNodeEdges(root, nodesDir, nodeID string) ([]specEdge, error) {
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

		if strings.HasSuffix(fname, ".geom.json") {
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

		if d, ok := edgefile.ReadEdgeDelta(root, nodeID, e.Label); ok {
			e.setDelta(d)
		}
		edges = append(edges, e)
	}
	return edges, nil
}
