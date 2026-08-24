package Topology

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"

	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgefile"
	SceneBuf "github.com/dtauraso/beadnetwork/Categories/Scene"
)

func LoadTree(root string) (TopoSpec, error) {
	var spec TopoSpec

	nodesDir := filepath.Join(root, "nodes")
	nodeDirs, err := readDirNames(nodesDir)
	if err != nil {
		return spec, fmt.Errorf("loadTree: list nodes dir %s: %w", nodesDir, err)
	}

	rowCount, err := validateNodeIDs(nodeDirs)
	if err != nil {
		return spec, err
	}
	spec.RowCount = rowCount

	constants, err := SceneBuf.LoadSceneConstants(root)
	if err != nil {
		return spec, err
	}
	spec.Constants = constants

	for _, nodeID := range nodeDirs {
		sn, err := loadNodeBase(root, nodesDir, nodeID)
		if err != nil {
			return spec, err
		}
		spec.Nodes = append(spec.Nodes, sn)

		edges, err := loadNodeEdges(root, nodesDir, nodeID)
		if err != nil {
			return spec, err
		}
		spec.Edges = append(spec.Edges, edges...)
	}

	ResolveEdgeDeltas(&spec)
	PlaceFromDeltas(&spec)

	reportEdgeClosure(&spec)

	ApplyDragOverlay(root, &spec)

	return spec, nil
}

func validateNodeIDs(nodeDirs []string) (int, error) {
	rowCount := 0
	seenIDs := make(map[int]string, len(nodeDirs))
	for _, name := range nodeDirs {
		n, err := strconv.Atoi(name)
		if err != nil {
			return 0, fmt.Errorf("loadTree: node directory %q is not a numeric id: %w", name, err)
		}
		if n < 1 {
			return 0, fmt.Errorf("loadTree: node directory %q has id %d, but node ids are 1-based (must be >= 1)", name, n)
		}
		if prev, dup := seenIDs[n]; dup {
			return 0, fmt.Errorf("loadTree: node directories %q and %q both parse to id %d — duplicate node id", prev, name, n)
		}
		seenIDs[n] = name
		if n > rowCount {
			rowCount = n
		}
	}
	return rowCount, nil
}

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
