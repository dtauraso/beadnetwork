// topo_spec.go — the topology spec format: the JSON shapes (specPort, specNode,
// specEdge, topoSpec, NodeData) and reading/validating a spec into memory
// (parseSpec, readSpec, validateNoFanIn), plus small per-node/per-edge helpers
// used at build time (label, toNodeGeom, broadcastBaseName, specPortsToGeom,
// nodeSendRule). loader.go's LoadTopology consumes this;
// build.go turns a parsed+validated topoSpec into the running graph.

package Wiring

import (
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"os"
)

// specNode mirrors the JSON node shape.
type specNode struct {
	ID    string    `json:"id"`
	Type  string    `json:"type"`
	Index *int      `json:"index,omitempty"`
	Data  *NodeData `json:"data,omitempty"`
	R     *float64  `json:"r,omitempty"` // optional per-node sphere radius for this node's edges (nil → default; see nodeR)
	// Scene polar (polar-model.md phase 2): the node's position as (r,θ,φ) about the scene
	// sphere. When present AND a persisted scene sphere exists, world = sceneCenter +
	// polar2cart(scenePolar) is AUTHORITATIVE over x/y/z (which stay for back-compat).
	ScenePolarR     *float64 `json:"scenePolarR,omitempty"`
	ScenePolarTheta *float64 `json:"scenePolarTheta,omitempty"`
	ScenePolarPhi   *float64 `json:"scenePolarPhi,omitempty"`
	// Quantized polar offset (quantized_layout.go): the node's (iTheta,iPhi,iR) integer
	// offset about the ONE scene sphere center — every node is independent (no reference/
	// parent). All three MUST be present together (all-or-nothing) for the stored offset
	// to be adopted; a node with any of these absent (an "old scene") is measured from its
	// scenePolar-derived world center instead (loader.go computeQuantizedLayout).
	QuantITheta *int `json:"quantITheta,omitempty"`
	QuantIPhi   *int `json:"quantIPhi,omitempty"`
	QuantIR     *int `json:"quantIR,omitempty"`
	// Per-node step constants (quantized_layout.go quantizedOffset.cTheta/cPhi/cR): this
	// node's OWN quantization step, turning its integer scalars into a world offset. nil
	// (unset) falls back to the global default (stepTheta/stepPhi/stepR).
	StepTheta *float64 `json:"stepTheta,omitempty"`
	StepPhi   *float64 `json:"stepPhi,omitempty"`
	StepR     *float64 `json:"stepR,omitempty"`
	// Gate marks this node as a two-neighbor GATE node (node_move.go).
	//
	// UNCONSUMED since the rule/gate/anchor cascade was deleted (2026-07-18): still
	// read/written for meta.json round-trip only (loader_tree.go copies it into
	// jsonMeta.Gate and back), but no code path branches on it. Do not assume it
	// drives behavior; grep call sites before relying on it again.
	Gate bool `json:"gate,omitempty"`
}

// label returns the node's human label: data.label when present and non-empty,
// otherwise the node id. Mirrors the TS `n.data?.label ?? n.id` fallback so the
// new-system label sidecar renders the same pill text the old spec store produced.
func (n specNode) label() string {
	if n.Data != nil && n.Data.Label != "" {
		return n.Data.Label
	}
	return n.ID
}

// toNodeGeom builds the geometry descriptor for edge-segment computation. A port
// contributes no geometry at all (docs/channels-not-ports.md — it is a load-time
// channel-binding ROLE, resolved by PortSpec/a.In()/a.Out() at build time, never here),
// so this no longer resolves or falls back to any port list.
func (n specNode) toNodeGeom(sceneCenter vec3) nodeGeom {
	// Position is POLAR (polar-frame-rewrite.md). The stored ScenePolar (r,θ,φ about the scene
	// sphere center) is the ONLY stored position and is adopted directly — there is no cartesian
	// x/y/z load path. When it is absent the node has no position (HasPos false → nodeWorldPos
	// returns origin). Scene presence does not gate polar adoption: the stored polar is
	// authoritative regardless.
	g := nodeGeom{nodeIdentity: nodeIdentity{Kind: n.Type, Label: n.label(), R: n.R, SceneCenter: sceneCenter}}
	if n.ScenePolarR != nil && n.ScenePolarTheta != nil && n.ScenePolarPhi != nil {
		g.ScenePolar = polar{R: *n.ScenePolarR, Theta: *n.ScenePolarTheta, Phi: *n.ScenePolarPhi}
		g.HasPos = true
	}
	return g
}

// broadcastBaseName strips a trailing digit suffix from a sourceHandle when the
// base name is an Broadcast port on the given kind, per kindBroadcastPorts (kind →
// set of Broadcast port names). e.g. "ToNext0" → "ToNext" for a kind with Broadcast
// port "ToNext". Returns the canonical port name and whether it resolved. Shared
// by buildFromSpec and validateSpec so the two normalizations can never drift.
func broadcastBaseName(handle, kind string, kindBroadcastPorts map[string]map[string]bool) (string, bool) {
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

// NodeData mirrors the JSON data block on a node.
type NodeData struct {
	// Label is the node's human label (optional). When absent, the node id is used
	// as the label. Streamed on node-geometry events for the new-system label sidecar.
	Label  string         `json:"label,omitempty"`
	Init   []int          `json:"init,omitempty"`
	Repeat bool           `json:"repeat,omitempty"`
	State  map[string]int `json:"state,omitempty"` // field-seeding: struct fields via wire:"data.state"
	// SendRules is the node-owned per-output-port send policy, keyed by output
	// port name (the sourceHandle, e.g. "ToNext0"). Absent ports default to
	// consumeGated. The send rule belongs to the SOURCE NODE, not the edge.
	SendRules map[string]string `json:"sendRules,omitempty"`
}

// specEdge mirrors the JSON edge shape.
// Fields tagged wire:"prop,..." are wire props emitted to wire-defs.ts by gen-node-defs.
//
// Source is NOT read from the file's own "source" key on disk — under the adjacency
// layout (topology/nodes/<id>/edges/<label>.json) the source is the directory the file
// sits in, and loadTree (loader_tree.go) fills Source in from that directory name after
// unmarshalling. The struct field still carries `json:"source"` so in-memory
// construction/tests can set it directly; it is simply redundant-and-unused on disk.
type specEdge struct {
	Label        string `json:"label"          wire:"prop,required,tsType:string"`
	Kind         string `json:"kind"`
	Source       string `json:"source"`
	SourceHandle string `json:"sourceHandle"`
	Target       string `json:"target"`
	TargetHandle string `json:"targetHandle"`
}

// topoSpec is the top-level JSON shape.
type topoSpec struct {
	Nodes []specNode `json:"nodes"`
	Edges []specEdge `json:"edges"`
	// RowCount is the buffer's row space: rows 0..RowCount-1, where row = id-1 (node identity
	// IS the row index, no id sidecar — see loader_tree.go's loadTree). It is the LARGEST node
	// id found in the tree, not len(Nodes): a deleted node leaves a gap row rather than
	// shifting later rows down. len(Nodes) <= RowCount always; equal only when the id space is
	// dense with no gaps.
	RowCount int
}

// WireRegistry maps edge label → *PacedWire. Each entry points to the wire owned by
// the destination port; multiple edges sharing a destination port map to the same *PacedWire.
// It is an internal build aid (binding source Out → wire); it is not returned.
type WireRegistry map[string]*wire.PacedWire

// parseSpec reads and parses the topology spec at path — a directory tree
// (loadTree; readSpec below rejects anything else) — into a topoSpec, WITHOUT
// validating or building. LoadTopology validates + builds from the result.
func parseSpec(path string) (topoSpec, error) {
	spec, err := readSpec(path)
	if err != nil {
		return topoSpec{}, err
	}
	// Fan-in is not part of the model: reject a spec where two edges target the same
	// input port, at parse time, so it fails cleanly at load rather than silently sharing
	// one wire deeper in the build.
	if err := validateNoFanIn(spec); err != nil {
		return topoSpec{}, fmt.Errorf("LoadTopology: %s: %w", path, err)
	}
	return spec, nil
}

// readSpec loads the raw topoSpec from the tree at path, without semantic validation
// (that is parseSpec's job).
// A topology is a DIRECTORY TREE and nothing else. The monolithic single-file form
// (a topology.json parsed straight into a topoSpec) is gone: two supported shapes meant
// every persister carried a second code path, and the tree is the form the editor writes,
// the form on disk, and the only one anything still produced.
//
// Fails loudly on a non-directory rather than falling back — a path that used to name a
// monolithic file would otherwise load as an empty spec and surface much later as a
// mystery empty scene.
func readSpec(path string) (topoSpec, error) {
	info, err := os.Stat(path) // path-resolution-ok: loader dispatch, not scene path resolution
	if err != nil {
		return topoSpec{}, fmt.Errorf("LoadTopology: stat %s: %w", path, err)
	}
	if !info.IsDir() { // path-resolution-ok: loader form check, not scene path resolution
		return topoSpec{}, fmt.Errorf("LoadTopology: %s is a file; a topology is a directory tree (nodes/<id>/{meta,data,inputs,outputs,edges}.json — adjacency layout). The monolithic single-file form was removed", path)
	}
	return loadTree(path)
}

// validateNoFanIn rejects a topology where two edges target the SAME destination input
// port (target + targetHandle). Fan-in was removed from the model: an input port accepts
// exactly ONE incident edge; multiple sources into one node use DISTINCT input ports — as
// every production node already does (e.g. a gate's FromLeft/FromRight, each fed by one
// edge). Enforced at parse so edge:wire:input-port is strictly 1:1 from load onward.
func validateNoFanIn(spec topoSpec) error {
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

// nodeSendRule looks up the node-owned per-output-port send rule for the
// given node id and output port name (sourceHandle). The rule lives on the
// SOURCE NODE's data.sendRules map, keyed by output port name. Ports not
// listed default to consumeGated.
func nodeSendRule(n specNode, port string) wire.SendRule {
	if n.Data == nil || n.Data.SendRules == nil {
		return wire.RuleConsumeGated
	}
	// ParseSendRule returns RuleConsumeGated for "" AND on error (unrecognised
	// value), so the fallback is already baked into its return value; the
	// error is deliberately ignored here (validate.go rejects bad values
	// before we reach here, so this is defence-in-depth only, and nodeSendRule's
	// callers aren't set up to handle a propagated error).
	rule, _ := wire.ParseSendRule(n.Data.SendRules[port])
	return rule
}
