package loadspec

import (
	"encoding/json"
	"fmt"
	"github.com/dtauraso/wirefold/PolarRules"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type JSONBase struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	IndexPhi   *int `json:"indexPhi,omitempty"`
	IndexTheta *int `json:"indexTheta,omitempty"`
	IndexR     *int `json:"indexR,omitempty"`

	Gate bool `json:"gate,omitempty"`

	Drag *PolarRules.DragRule `json:"drag,omitempty"`

	SelfDrag *PolarRules.DragRule `json:"selfDrag,omitempty"`

	TopTiltVectorPhiIdx *int32 `json:"topTiltVectorThetaIdx,omitempty"`
}

func loadNodeBase(root, nodesDir, nodeID string) (specNode, error) {
	nodeDir := filepath.Join(nodesDir, nodeID)

	basePath := filepath.Join(nodeDir, "base.json")
	baseRaw, err := os.ReadFile(basePath)
	if err != nil {
		return specNode{}, fmt.Errorf("loadTree: node %q base: %w", nodeID, err)
	}
	var base JSONBase
	if err := json.Unmarshal(baseRaw, &base); err != nil {
		return specNode{}, fmt.Errorf("loadTree: node %q base parse: %w", nodeID, err)
	}

	sn := specNode{
		ID:                  base.ID,
		Type:                base.Type,
		IndexPhi:            base.IndexPhi,
		IndexTheta:          base.IndexTheta,
		IndexR:              base.IndexR,
		Gate:                base.Gate,
		Drag:                base.Drag,
		SelfDrag:            base.SelfDrag,
		TopTiltVectorPhiIdx: base.TopTiltVectorPhiIdx,
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

		edges = append(edges, e)
	}
	return edges, nil
}
