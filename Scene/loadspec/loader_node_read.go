package loadspec

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/Node/Edge/edgefile"
	"github.com/dtauraso/wirefold/Node/nodefile"
)

func readLeaf(path string, v any) bool {
	return ReadIfExists(path, v)
}

func readOptInt(path string) *int {
	var v int
	if !readLeaf(path, &v) {
		return nil
	}
	return &v
}

func readOptInt32(path string) *int32 {
	var v int32
	if !readLeaf(path, &v) {
		return nil
	}
	return &v
}

func loadNodeBase(root, nodesDir, nodeID string) (Node, error) {
	nodeDir := filepath.Join(nodesDir, nodeID)

	baseDir := nodefile.BaseDir(nodeDir)
	var nodeType string
	if !readLeaf(filepath.Join(baseDir, nodefile.FileType), &nodeType) {
		return Node{}, fmt.Errorf("loadTree: node %q has no %s", nodeID, nodefile.FileType)
	}

	sn := Node{
		ID:                  nodeID,
		Type:                nodeType,
		IndexPhi:            readOptInt(filepath.Join(baseDir, nodefile.FileIndexPhi)),
		IndexTheta:          readOptInt(filepath.Join(baseDir, nodefile.FileIndexTheta)),
		IndexR:              readOptInt(filepath.Join(baseDir, nodefile.FileIndexR)),
		Drag:                nodefile.ReadDragRule(filepath.Join(baseDir, nodefile.DirDragRule)),
		SelfDrag:            nodefile.ReadDragRule(filepath.Join(baseDir, nodefile.DirSelfRule)),
		TopTiltVectorPhiIdx: readOptInt32(filepath.Join(baseDir, nodefile.FileTiltIdx)),
	}
	readLeaf(filepath.Join(baseDir, nodefile.FileGate), &sn.Gate)

	nd, err := readNodeData(nodeDir)
	if err != nil {
		return Node{}, fmt.Errorf("loadTree: node %q data: %w", nodeID, err)
	}
	sn.Data = nd

	return sn, nil
}

func loadNodeEdges(root, nodesDir, nodeID string) ([]Edge, error) {
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

	var edges []Edge
	for _, label := range edgeFiles {
		edgeDir := filepath.Join(edgesDir, label)
		if st, err := os.Stat(edgeDir); err != nil || !st.IsDir() { // path-resolution-ok: an edge is a directory; skipping strays under edges/
			continue
		}
		e := Edge{Label: label, Source: nodeID}
		readLeaf(filepath.Join(edgeDir, edgefile.FileKind), &e.Kind)
		readLeaf(filepath.Join(edgeDir, edgefile.FileSourceHandle), &e.SourceHandle)
		readLeaf(filepath.Join(edgeDir, edgefile.FileTarget), &e.Target)
		readLeaf(filepath.Join(edgeDir, edgefile.FileTargetHandle), &e.TargetHandle)
		e.DeltaIndexR = readOptInt(filepath.Join(edgeDir, edgefile.FileDeltaIndexR))
		e.DeltaIndexPhi = readOptInt(filepath.Join(edgeDir, edgefile.FileDeltaIndexPhi))
		e.DeltaIndexTheta = readOptInt(filepath.Join(edgeDir, edgefile.FileDeltaIndexTheta))

		edges = append(edges, e)
	}
	return edges, nil
}
