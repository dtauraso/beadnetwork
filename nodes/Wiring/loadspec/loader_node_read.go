package loadspec

import (
	"encoding/json"
	"fmt"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodefile"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func readJSONFile(path string, v any) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return json.Unmarshal(raw, v) == nil
}

func readOptInt(path string) *int {
	var v int
	if !readJSONFile(path, &v) {
		return nil
	}
	return &v
}

func readOptInt32(path string) *int32 {
	var v int32
	if !readJSONFile(path, &v) {
		return nil
	}
	return &v
}

func loadNodeBase(root, nodesDir, nodeID string) (specNode, error) {
	nodeDir := filepath.Join(nodesDir, nodeID)

	baseDir := nodefile.BaseDir(nodeDir)
	var nodeType string
	if !readJSONFile(filepath.Join(baseDir, nodefile.FileType), &nodeType) {
		return specNode{}, fmt.Errorf("loadTree: node %q has no %s", nodeID, nodefile.FileType)
	}

	sn := specNode{
		ID:                  nodeID,
		Type:                nodeType,
		IndexPhi:            readOptInt(filepath.Join(baseDir, nodefile.FileIndexPhi)),
		IndexTheta:          readOptInt(filepath.Join(baseDir, nodefile.FileIndexTheta)),
		IndexR:              readOptInt(filepath.Join(baseDir, nodefile.FileIndexR)),
		Drag:                nodefile.ReadDragRule(filepath.Join(baseDir, nodefile.DirDragRule)),
		SelfDrag:            nodefile.ReadDragRule(filepath.Join(baseDir, nodefile.DirSelfRule)),
		TopTiltVectorPhiIdx: readOptInt32(filepath.Join(baseDir, nodefile.FileTiltIdx)),
	}
	readJSONFile(filepath.Join(baseDir, nodefile.FileGate), &sn.Gate)

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
