// topo_spec.go — the topology spec format: the JSON shapes (specPort, specNode,
// specEdge, topoSpec, NodeData) and reading/validating a spec into memory
// (parseSpec, readSpec, validateNoFanIn), plus small per-node/per-edge helpers
// used at build time (label, toNodeGeom, broadcastBaseName, specPortsToGeom,
// nodeSendRule). loader.go's LoadTopology/LoadTopologyFromJSON consume this;
// build.go turns a parsed+validated topoSpec into the running graph.

package Wiring

import (
	"encoding/json"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"os"
)

// specPort mirrors the per-node inputs/outputs entries in topology.json.
// AnchorId is the only placement field; side/slot/anchor have been removed.
type specPort struct {
	Name     string   `json:"name"`
	AnchorId *int     `json:"anchorId,omitempty"` // optional ring-anchor index (flat array); highest priority
	PortR    *float64 `json:"portR,omitempty"`    // optional per-port radius (distance from node center); nil → nodeRadius(kind) fallback
}

// specNode mirrors the JSON node shape.
type specNode struct {
	ID      string     `json:"id"`
	Type    string     `json:"type"`
	Index   *int       `json:"index,omitempty"`
	Data    *NodeData  `json:"data,omitempty"`
	Inputs  []specPort `json:"inputs,omitempty"`
	Outputs []specPort `json:"outputs,omitempty"`
	R       *float64   `json:"r,omitempty"` // optional per-node sphere radius for this node's edges (nil → default; see nodeR)
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
	// LocalPolars is this node's list of per-neighbor local polars (layout_holder.go
	// LocalPolar) — one per domain double-link this node is an endpoint of, measured
	// with ITSELF as center. Absent (nil) → computed fresh at load (computeLocalPolars).
	LocalPolars []specLocalPolar `json:"localPolars,omitempty"`
	// LocalPoleTheta/LocalPolePhi is the measurement pole (rotating_pole.go localPole
	// result) that LocalPolars' stored indices were last quantized about
	// (memory/feedback_abc_times_constant_not_rederive.md: carry the stored pole forward
	// rather than re-deriving it). Both absent → no stored pole; computeLocalPolars falls
	// back to recomputing the pole fresh from live composed centers (the load-time
	// cart↔polar boundary, unchanged by this).
	LocalPoleTheta *float64 `json:"localPoleTheta,omitempty"`
	LocalPolePhi   *float64 `json:"localPolePhi,omitempty"`
	// CascadeEdges is this node's STORED list of cascade-neighbor ids (nodes/<id>/
	// cascade-edges.json, loader_tree.go) — the sole source of truth for delta-forward
	// propagation (nodeMover.forwardDelta) and the cascade-link overlay's rendered pairs.
	// NOT derived from LocalPolars/the domain-edge adjacency at load (the computed
	// cascade_links.go machinery was removed) — hand-authored/persisted data. Absent file
	// → nil (empty list), matching every other per-node optional-file convention here.
	CascadeEdges []string `json:"cascadeEdges,omitempty"`
	// CascadeKinds maps each CascadeEdges neighbor id → that neighbor's kind name, stored
	// in the same cascade-edges.json file so a node's cascade channels carry the peer kind
	// directly (no central id→kind lookup at load). Consumed by nodeMover.forwardDelta for
	// kind-selective delta routing (see that method). Absent → nil.
	CascadeKinds map[string]string `json:"cascadeKinds,omitempty"`
	// Gate marks this node as a two-neighbor GATE node (node_move.go): on a direct
	// drag it solves its own equal-radii landing position against its two domain
	// neighbors (derived from LocalPolars, in the same order), commits, and
	// self-triggers its own edge-c equalize. NOT derivable from degree (other
	// 2-link nodes exist that are plain leaves) — authored in the spec.
	//
	// UNCONSUMED since the rule/gate/anchor cascade was deleted (2026-07-18): still
	// read/written for meta.json round-trip only (loader_tree.go copies it into
	// jsonMeta.Gate and back), but no code path branches on it. Do not assume it
	// drives behavior; grep call sites before relying on it again.
	Gate bool `json:"gate,omitempty"`
}

// specLocalPolar mirrors one entry of a node's persisted localPolars list
// (loader_tree.go jsonMeta.LocalPolars carries the same shape).
type specLocalPolar struct {
	To          string  `json:"to"`
	QuantITheta int     `json:"quantITheta"`
	QuantIPhi   int     `json:"quantIPhi"`
	QuantIR     int     `json:"quantIR"`
	StepTheta   float64 `json:"stepTheta,omitempty"`
	StepPhi     float64 `json:"stepPhi,omitempty"`
	StepR       float64 `json:"stepR,omitempty"`
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

// toNodeGeom builds the geometry descriptor for arc-length computation,
// resolving the port lists from the spec node (falling back to the kind's
// registry ports with default sides when the spec omits inputs/outputs).
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
	g.Inputs = specPortsToGeom(n.Inputs)
	g.Outputs = specPortsToGeom(n.Outputs)
	// Fallback to registry ports when the spec omits the lists (keeps geometry
	// well-defined for hand-written topologies that rely on default placement).
	if len(g.Inputs) == 0 || len(g.Outputs) == 0 {
		if bind, ok := Registry[n.Type]; ok {
			if len(g.Inputs) == 0 {
				for _, p := range bind.Ports {
					if p.Dir == PortIn {
						g.Inputs = append(g.Inputs, portGeom{Name: p.Name})
					}
				}
			}
			if len(g.Outputs) == 0 {
				for _, p := range bind.Ports {
					if p.Dir == PortOut || p.Dir == PortBroadcast {
						g.Outputs = append(g.Outputs, portGeom{Name: p.Name})
					}
				}
			}
		}
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

func specPortsToGeom(ports []specPort) []portGeom {
	out := make([]portGeom, 0, len(ports))
	for _, p := range ports {
		out = append(out, portGeom(p))
	}
	return out
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
}

// WireRegistry maps edge label → *PacedWire. Each entry points to the wire owned by
// the destination port; multiple edges sharing a destination port map to the same *PacedWire.
// It is an internal build aid (binding source Out → wire); it is not returned.
type WireRegistry map[string]*wire.PacedWire

// parseSpec reads and parses the topology spec at path — a directory tree
// (loadTree) or a monolithic topology.json — into a topoSpec, WITHOUT validating
// or building. LoadTopology validates + builds from the result.
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

// readSpec loads the raw topoSpec from either a directory tree or a single JSON file,
// without semantic validation (that is parseSpec's job).
func readSpec(path string) (topoSpec, error) {
	if info, err := os.Stat(path); err == nil && info.IsDir() { // path-resolution-ok: loader dispatch, not scene path resolution
		return loadTree(path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return topoSpec{}, fmt.Errorf("LoadTopology: read %s: %w", path, err)
	}
	var spec topoSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return topoSpec{}, fmt.Errorf("LoadTopology: parse %s: %w", path, err)
	}
	return spec, nil
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
