package loadspec

import (
	"encoding/json"
	"fmt"
	"github.com/dtauraso/wirefold/src/Node/Wiring/edgefile"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodefile"
	"os"
	"path/filepath"
	"sort"
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
	for _, label := range edgeFiles {
		edgeDir := filepath.Join(edgesDir, label)
		if st, err := os.Stat(edgeDir); err != nil || !st.IsDir() { // path-resolution-ok: an edge is a directory; skipping strays under edges/
			continue
		}
		e := specEdge{Label: label, Source: nodeID}
		readJSONFile(filepath.Join(edgeDir, edgefile.FileKind), &e.Kind)
		readJSONFile(filepath.Join(edgeDir, edgefile.FileSourceHandle), &e.SourceHandle)
		readJSONFile(filepath.Join(edgeDir, edgefile.FileTarget), &e.Target)
		readJSONFile(filepath.Join(edgeDir, edgefile.FileTargetHandle), &e.TargetHandle)
		e.DeltaIndexR = readOptInt(filepath.Join(edgeDir, edgefile.FileDeltaIndexR))
		e.DeltaIndexPhi = readOptInt(filepath.Join(edgeDir, edgefile.FileDeltaIndexPhi))
		e.DeltaIndexTheta = readOptInt(filepath.Join(edgeDir, edgefile.FileDeltaIndexTheta))

		edges = append(edges, e)
	}
	return edges, nil
}
