// build.go — turns an already-parsed and validated topoSpec into the running
// graph: node geometry, quantized layout, wire allocation, the MoveDispatch,
// and the built []wire.Node themselves. buildFromSpec orchestrates the phase
// helpers (each a method on buildCtx, split into its own file by phase — see
// build_geometry.go, build_wires.go, build_move_dispatch.go, build_edge_maps.go,
// build_nodes.go, plus loader_layout.go's computeQuantizedLayout/computeReachRadii)
// in the same order the original monolithic loader.go function performed them;
// behavior is unchanged. loader.go's LoadTopology calls buildFromSpec after
// parsing + validating via topo_spec.go.

package Wiring

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// buildCtx carries the shared state threaded through the buildFromSpec phase
// helpers below. Each phase method populates its own fields from spec (and
// fields written by earlier phases); buildFromSpec calls them in order and
// stays a short orchestrator. Splitting on struct fields (rather than
// threading a long parameter list) mirrors the original function's data
// flow exactly — no behavior changes, only the grouping into named steps.
type buildCtx struct {
	ctx      context.Context
	spec     topoSpec
	tr       *T.Trace
	clk      clock.Clock
	sphere   geom.SceneSphere
	hasScene bool
	// scenePath is the tree being loaded — carried so the build can ask which DRAG this
	// scene uses (scene/scene_capabilities.go's SceneUsesQuantizedDrag). It is the loaded scene's own
	// path, never the anchor: the loader knows which tree it is opening, not which tab
	// pointed it there. "" (test call sites that build a spec directly) reads as unknown,
	// which SceneUsesQuantizedDrag answers with the quantized drag every scene had before
	// scenes were selectable.
	scenePath string

	// Phase 1: node geometry + world centers.
	nodeGeoms map[string]nodegeom.NodeGeom
	centers   map[string]vec3

	// Phase 1b: quantized flat absolute scene-polar layout (quantized_layout.go) —
	// resolved BEFORE reach/wire/dispatch phases so every later phase computes from the
	// COMPOSED (authoritative) centers, not the raw loaded ones. Every node is a root
	// measured about the scene center — no reference/parent concept.
	quantizedOffsets map[string]quantizedOffset

	// Phase 4: per-destination-port wire allocation + per-edge geometry.
	destWire      map[string]*wire.PacedWire
	edgeWire      WireRegistry
	edgeEndpoints map[string]inputcodec.EdgeEndpoints
	// edgeSteps is each edge's own bead-step count (docs/bead-model/bead-lattice.md "The
	// count") — its INITIAL published value, computed once at load time from the
	// source node's own b.localPolars entry to the target. The source node's own
	// goroutine recomputes and republishes this same integer every cycle once
	// running (chain_beads.go's chainBeads); this load-time value is what the
	// wire's dwell and the first-frame chain layout start from before that.
	edgeSteps    map[string]int
	edgeSegments map[string]wireSegment

	// Phase 5: the MoveDispatch.
	md *MoveDispatch

	// speedSinks accumulates the SEND end of every speed channel created for
	// any clock-owning goroutine across the whole build — edge movers
	// (buildMoveDispatch, via newMoveDispatch) and per-node goroutines
	// (buildNodes, via each kind's own builder calling a.SpeedCh()). Returned by
	// buildFromSpec/LoadTopology (per-goroutine-clock.md "Delivery").
	speedSinks []chan float64

	// Phase 6: id→type map and per-kind Broadcast port set.
	nodeType           map[string]string
	kindBroadcastPorts map[string]map[string]bool

	// Phase 7: inbound/outbound edge maps.
	inbound        map[string]map[string]string
	outbound       map[string]map[string][]string
	outboundHandle map[string]map[string][]string

	// Phase 8: built nodes + the paced-Out sink.
	outSink map[string]*wire.Out
	nodes   []wire.Node

	// vectorOutByNode/vectorInByNode: node id -> its own dedicated tilt-vector
	// channel end (tilt_vector_channel.go), built once by allocateVectorChannels
	// for every edge whose BOTH endpoint kinds asked for one (today: PairNode
	// only). A node id absent from a map has no vector channel on that side.
	vectorOutByNode map[string]chan tiltvector.TiltVectorMsg
	vectorInByNode  map[string]chan tiltvector.TiltVectorMsg
}

// buildFromSpec constructs nodes, wires, and the MoveDispatch from an already-parsed
// and validated topoSpec. It orchestrates the phase helpers below in the same order
// the original monolithic function performed them; behavior is unchanged.
func buildFromSpec(ctx context.Context, spec topoSpec, tr *T.Trace, clk clock.Clock, sphere geom.SceneSphere, hasScene bool, scenePath string) ([]wire.Node, inputcodec.SlotRegistry, *MoveDispatch, []chan float64, error) {
	b := &buildCtx{ctx: ctx, spec: spec, tr: tr, clk: clk, sphere: sphere, hasScene: hasScene, scenePath: scenePath}

	b.computeNodeGeometry()
	b.computeQuantizedLayout()
	b.computeReachRadii()
	b.allocateWires()
	b.allocateVectorChannels()
	if err := b.buildMoveDispatch(); err != nil {
		return nil, nil, nil, nil, err
	}
	b.buildTypeMaps()
	b.buildEdgeMaps()
	if err := b.buildNodes(); err != nil {
		return nil, nil, nil, nil, err
	}
	// finalizeActors runs AFTER buildNodes: that is when every kind's own build func has
	// run, so every BuildArgs.ClaimSelfDrive call (PairNode, the pair scene) has
	// already recorded itself in md.selfDriveClaimed. Only NOW is it known which node
	// ids get a real nodeMover actor at all (task/pair-node-owns-itself).
	b.md.finalizeActors(&b.speedSinks)
	b.bindDispatch()

	return b.nodes, inputcodec.SlotRegistry(b.destWire), b.md, b.speedSinks, nil
}
