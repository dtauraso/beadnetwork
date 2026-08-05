// loader.go — runtime topology loader entry points.
//
// LoadTopology reads the topology directory tree, allocates one PacedWire per destination
// port, and returns ([]Node, SlotRegistry, *MoveDispatch).
// An edge-label-keyed WireRegistry is built internally to bind each source Out to
// its wire, but it is not returned: no caller consumed the map.
//
// The spec format (specNode/specEdge/topoSpec/NodeData) and reading/validating it
// (parseSpec/readSpec/validateNoFanIn) live in topo_spec.go. Turning a parsed spec
// into the running graph (buildCtx/buildFromSpec and its phase helpers) lives in
// build.go. This file holds only the two public entry points.
//
// Key behaviors:
//   - One *PacedWire per (destNode, destPort), fed by exactly one edge — fan-in
//     (two edges into one port) is rejected at parse (validateNoFanIn), so a
//     destination port maps 1:1 to a wire and a single incident edge.
//   - SlotRegistry maps "target.targetHandle" → wire for create/delete ops.
//   - Input nodes: data.init values pre-seeded via pw.Send in a goroutine.
//   - Time: data.state["held"] → Held via wire:"data.state" tag.
//   - Slice output ports (ToEdge): all outbound wires appended in spec order.
//   - Output ports with no outbound edge: dead-end chan int (buf 1).

package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// LoadTopology reads the JSON file at jsonPath and constructs []Node plus a
// SlotRegistry (keyed by "target.targetHandle" for delivery acks) and a MoveDispatch
// (key→inbox registry for the decentralized node-move path: each node and edge owns
// its own recompute).
//
// clk is the single monotonic clock injected into every PacedWire so each wire
// times its own delivery on it (MODEL.md: exactly one clock). Production and
// tests alike pass a RealClock — the model is sleep-only.
//
// The 5th return value is the build-wide list of SEND ends of every speed
// channel created for a clock-owning goroutine (per-goroutine-clock.md
// "Delivery") — one per goroutine, collected ONCE here at load time before any
// goroutine spawns, so the set never needs a lock: it is written only during
// this call and read thereafter by exactly one goroutine (stdin_reader's),
// never touched by the goroutines that own the receive ends. Most callers
// (tests that don't drive a speed slider) can discard it.
func LoadTopology(ctx context.Context, jsonPath string, tr *T.Trace, clk wire.Clock) ([]wire.Node, SlotRegistry, *MoveDispatch, []chan float64, error) {
	BuildRegistry()
	spec, err := parseSpec(jsonPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if err := validateSpec(&spec); err != nil {
		return nil, nil, nil, nil, err
	}
	// Load the persisted scene sphere (if any) BEFORE positioning nodes, so nodes stored as
	// scene polar can be placed as sceneCenter + polar2cart(scenePolar). A persisted sphere
	// is not derived from node positions, so there is no circularity; a fresh/legacy scene
	// has none and nodes fall back to cartesian x/y/z (polar-model.md phase 2b).
	sphere, hasScene := loadSceneSphere(jsonPath)
	return buildFromSpec(ctx, spec, tr, clk, sphere, hasScene, jsonPath)
}
