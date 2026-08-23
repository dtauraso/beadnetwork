package loadspec

import (
	"fmt"
	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
	"os"
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/Categories/Node/Edge/edgefile"
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

	baseDir := NodeBuf.BaseDir(nodeDir)
	var nodeType string
	if !readLeaf(filepath.Join(baseDir, NodeBuf.FileType), &nodeType) {
		return Node{}, fmt.Errorf("loadTree: node %q has no %s", nodeID, NodeBuf.FileType)
	}

	sn := Node{
		ID:                  nodeID,
		Type:                nodeType,
		IndexPhi:            readOptInt(filepath.Join(baseDir, NodeBuf.FileIndexPhi)),
		IndexTheta:          readOptInt(filepath.Join(baseDir, NodeBuf.FileIndexTheta)),
		IndexR:              readOptInt(filepath.Join(baseDir, NodeBuf.FileIndexR)),
		Drag:                NodeBuf.ReadDragRule(filepath.Join(baseDir, NodeBuf.DirDragRule)),
		SelfDrag:            NodeBuf.ReadDragRule(filepath.Join(baseDir, NodeBuf.DirSelfRule)),
		TopTiltVectorPhiIdx: readOptInt32(filepath.Join(baseDir, NodeBuf.FileTiltIdx)),
	}
	readLeaf(filepath.Join(baseDir, NodeBuf.FileGate), &sn.Gate)

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
