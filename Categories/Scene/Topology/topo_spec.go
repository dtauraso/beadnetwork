package Topology

import (
	"fmt"
	"os"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/PolarRulesPanel"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Node struct {
	ID   string
	Type string
	Data *NodeBuf.NodeData

	IndexPhi   *int
	IndexTheta *int
	IndexR     *int

	DragIndexPhi   *int
	DragIndexTheta *int
	DragIndexR     *int

	Gate bool

	Drag *PolarRulesPanel.DragRule

	SelfDrag *PolarRulesPanel.DragRule

	TopTiltVectorPhiIdx *int32
}

func (n Node) Label() string {
	if n.Data != nil && n.Data.Label != "" {
		return n.Data.Label
	}
	return n.ID
}

type NodeData struct {
	Label  string
	Init   []int
	Repeat bool
	State  map[string]int

	SendRules map[string]string
}

type Edge struct {
	Label        string `wire:"prop,required,tsType:string"`
	Kind         string
	Source       string
	SourceHandle string
	Target       string
	TargetHandle string

	DeltaIndexR     *int
	DeltaIndexPhi   *int
	DeltaIndexTheta *int

	DragDeltaIndexR     *int
	DragDeltaIndexPhi   *int
	DragDeltaIndexTheta *int
}

type TopoSpec struct {
	Nodes []Node
	Edges []Edge

	RowCount int

	Constants polarindex.SceneConstants
}

func ParseSpec(path string) (TopoSpec, error) {
	spec, err := readSpec(path)
	if err != nil {
		return TopoSpec{}, err
	}

	if err := validateNoFanIn(spec); err != nil {
		return TopoSpec{}, fmt.Errorf("LoadTopology: %s: %w", path, err)
	}
	return spec, nil
}

func readSpec(path string) (TopoSpec, error) {
	info, err := os.Stat(path) // path-resolution-ok: loader dispatch, not scene path resolution
	if err != nil {
		return TopoSpec{}, fmt.Errorf("LoadTopology: stat %s: %w", path, err)
	}
	if !info.IsDir() { // path-resolution-ok: loader form check, not scene path resolution
		return TopoSpec{}, fmt.Errorf("LoadTopology: %s is a file; a topology is a directory tree (nodes/<id>/{meta,data,inputs,outputs,edges}.json — adjacency layout). The monolithic single-file form was removed", path)
	}
	return LoadTree(path)
}

func validateNoFanIn(spec TopoSpec) error {
	seen := make(map[string]string, len(spec.Edges))
	for _, e := range spec.Edges {
		key := e.Target + "." + e.TargetHandle
		if prev, dup := seen[key]; dup {
			return fmt.Errorf("fan-in not allowed: edges %q and %q both target input port %s.%s — an input port takes exactly one edge; use distinct input ports for multiple sources",
				prev, e.Label, e.Target, e.TargetHandle)
		}
		seen[key] = e.Label
	}
	return nil
}
