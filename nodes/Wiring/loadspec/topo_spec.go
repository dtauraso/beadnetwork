package loadspec

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	"github.com/dtauraso/wirefold/nodes/spatial"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

type specNode struct {
	ID   string    `json:"id"`
	Type string    `json:"type"`
	Data *NodeData `json:"data,omitempty"`

	IndexPhi   *int `json:"indexPhi,omitempty"`
	IndexTheta *int `json:"indexTheta,omitempty"`
	IndexR     *int `json:"indexR,omitempty"`

	DragIndexPhi   *int `json:"-"`
	DragIndexTheta *int `json:"-"`
	DragIndexR     *int `json:"-"`

	Gate bool `json:"gate,omitempty"`

	Drag *polar.DragRule `json:"drag,omitempty"`

	SelfDrag *polar.DragRule `json:"selfDrag,omitempty"`

	TopTiltVectorPhiIdx *int32 `json:"topTiltVectorThetaIdx,omitempty"`
}

func (n specNode) label() string {
	if n.Data != nil && n.Data.Label != "" {
		return n.Data.Label
	}
	return n.ID
}

func (n specNode) ToNodeGeom(sceneCenter spatial.Vec3, sc polarindex.SceneConstants) nodegeom.NodeGeom {

	g := nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: n.Type, Label: n.label(), SceneCenter: sceneCenter, SceneConstants: sc}}
	if n.hasPoint() {
		g.BaseIndex = n.index()
		g.HasPos = true
	}
	if n.DragIndexPhi != nil && n.DragIndexTheta != nil && n.DragIndexR != nil {
		g.DragIndex = polarindex.Offset{Phi: *n.DragIndexPhi, Theta: *n.DragIndexTheta, R: *n.DragIndexR}
	}
	return g
}

func BroadcastBaseName(handle, kind string, kindBroadcastPorts map[string]map[string]bool) (string, bool) {
	if len(handle) == 0 {
		return handle, false
	}
	last := handle[len(handle)-1]
	if last < '0' || last > '9' {
		return handle, false
	}
	base := handle[:len(handle)-1]
	if kindBroadcastPorts[kind][base] {
		return base, true
	}
	return handle, false
}

type NodeData struct {
	Label  string         `json:"label,omitempty"`
	Init   []int          `json:"init,omitempty"`
	Repeat bool           `json:"repeat,omitempty"`
	State  map[string]int `json:"state,omitempty"`

	SendRules map[string]string `json:"sendRules,omitempty"`
}

type specEdge struct {
	Label        string `json:"label"          wire:"prop,required,tsType:string"`
	Kind         string `json:"kind"`
	Source       string `json:"source"`
	SourceHandle string `json:"sourceHandle"`
	Target       string `json:"target"`
	TargetHandle string `json:"targetHandle"`

	// pole convention as a node's absolute index. See edge_delta.go: A + D = B.
	DeltaIndexR     *int `json:"deltaIndexR,omitempty"`
	DeltaIndexPhi   *int `json:"deltaIndexPhi,omitempty"`
	DeltaIndexTheta *int `json:"deltaIndexTheta,omitempty"`

	DragDeltaIndexR     *int `json:"-"`
	DragDeltaIndexPhi   *int `json:"-"`
	DragDeltaIndexTheta *int `json:"-"`
}

type TopoSpec struct {
	Nodes []specNode `json:"nodes"`
	Edges []specEdge `json:"edges"`

	RowCount int

	Constants polarindex.SceneConstants
}

type WireRegistry map[string]*wire.PacedWire

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

func NodeSendRule(n specNode, port string) outport.SendRule {
	if n.Data == nil || n.Data.SendRules == nil {
		return outport.RuleConsumeGated
	}

	rule, _ := outport.ParseSendRule(n.Data.SendRules[port])
	return rule
}
