package loadspec

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

type specNode struct {
	ID    string    `json:"id"`
	Type  string    `json:"type"`
	Index *int      `json:"index,omitempty"`
	Data  *NodeData `json:"data,omitempty"`
	R     *float64  `json:"r,omitempty"`

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

	TopTiltVectorThetaIdx *int32 `json:"topTiltVectorThetaIdx,omitempty"`
}

func (n specNode) label() string {
	if n.Data != nil && n.Data.Label != "" {
		return n.Data.Label
	}
	return n.ID
}

func (n specNode) ToNodeGeom(sceneCenter wire.Vec3) nodegeom.NodeGeom {

	g := nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: n.Type, Label: n.label(), R: n.R, SceneCenter: sceneCenter}}
	if n.ScenePolarR != nil && n.ScenePolarTheta != nil && n.ScenePolarPhi != nil {
		g.ScenePolar = geom.Polar{R: *n.ScenePolarR, Theta: *n.ScenePolarTheta, Phi: *n.ScenePolarPhi}
		g.HasPos = true
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
}

type TopoSpec struct {
	Nodes []specNode `json:"nodes"`
	Edges []specEdge `json:"edges"`

	RowCount int
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
