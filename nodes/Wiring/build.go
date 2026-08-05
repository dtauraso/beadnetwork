// build.go — turns an already-parsed and validated topoSpec into the running
// graph: node geometry, quantized layout, wire allocation, the MoveDispatch,
// and the built []wire.Node themselves. buildFromSpec orchestrates the phase
// helpers below (all methods on buildCtx) in the same order the original
// monolithic loader.go function performed them; behavior is unchanged.
// loader.go's LoadTopology calls buildFromSpec after parsing + validating via
// topo_spec.go.

package Wiring

import (
	"context"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"

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
	clk      wire.Clock
	sphere   sceneSphere
	hasScene bool
	// scenePath is the tree being loaded — carried so the build can ask which DRAG this
	// scene uses (scene_tabs.go's SceneUsesQuantizedDrag). It is the loaded scene's own
	// path, never the anchor: the loader knows which tree it is opening, not which tab
	// pointed it there. "" (test call sites that build a spec directly) reads as unknown,
	// which SceneUsesQuantizedDrag answers with the quantized drag every scene had before
	// scenes were selectable.
	scenePath string

	// Phase 1: node geometry + world centers.
	nodeGeoms map[string]nodeGeom
	centers   map[string]vec3

	// Phase 1b: quantized flat absolute scene-polar layout (quantized_layout.go) —
	// resolved BEFORE reach/wire/dispatch phases so every later phase computes from the
	// COMPOSED (authoritative) centers, not the raw loaded ones. Every node is a root
	// measured about the scene center — no reference/parent concept.
	quantizedOffsets map[string]quantizedOffset

	// Phase 4: per-destination-port wire allocation + per-edge geometry.
	destWire      map[string]*wire.PacedWire
	edgeWire      WireRegistry
	edgeEndpoints map[string]EdgeEndpoints
	// edgeSteps is each edge's own bead-step count (docs/bead-lattice.md "The
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
}

// buildFromSpec constructs nodes, wires, and the MoveDispatch from an already-parsed
// and validated topoSpec. It orchestrates the phase helpers below in the same order
// the original monolithic function performed them; behavior is unchanged.
func buildFromSpec(ctx context.Context, spec topoSpec, tr *T.Trace, clk wire.Clock, sphere sceneSphere, hasScene bool, scenePath string) ([]wire.Node, SlotRegistry, *MoveDispatch, []chan float64, error) {
	b := &buildCtx{ctx: ctx, spec: spec, tr: tr, clk: clk, sphere: sphere, hasScene: hasScene, scenePath: scenePath}

	b.computeNodeGeometry()
	b.computeQuantizedLayout()
	b.computeReachRadii()
	b.allocateWires()
	if err := b.buildMoveDispatch(); err != nil {
		return nil, nil, nil, nil, err
	}
	b.buildTypeMaps()
	b.buildEdgeMaps()
	if err := b.buildNodes(); err != nil {
		return nil, nil, nil, nil, err
	}
	b.bindDispatch()

	return b.nodes, SlotRegistry(b.destWire), b.md, b.speedSinks, nil
}

// computeNodeGeometry builds the id→geometry map used for arc-length computation
// at wire construction (nodeGeom carries kind/dims/port side+slot so the Go arc
// length mirrors buildPortCurve exactly), plus the shared world-center map built
// ONCE from that geometry and reused by the reach-radius pass, the aimed-port
// registry, and the edge-geometry centerOf closure. Each node's world center is
// loaded directly from its spec (meta.json x/y/z, injected as nodeGeom.Center in
// toNodeGeom); nothing later mutates a node's Center, so this snapshot stays
// authoritative for the whole build (the reach-radius pass writes ReachR, which
// does not affect centers).
func (b *buildCtx) computeNodeGeometry() {
	nodeGeoms := map[string]nodeGeom{}
	for _, n := range b.spec.Nodes {
		nodeGeoms[n.ID] = n.toNodeGeom(b.sphere.Center)
	}
	b.nodeGeoms = nodeGeoms

	centers := map[string]vec3{}
	for id, g := range nodeGeoms {
		if g.HasPos {
			centers[id] = nodeWorldPos(g)
		}
	}
	b.centers = centers
}

// computeQuantizedLayout makes the quantized flat absolute scene-polar layout
// (quantized_layout.go) AUTHORITATIVE for every node's world center. It resolves each
// allocateWires allocates one *PacedWire per destination port (one edge per port —
// fan-in is rejected at parse) and computes each edge's own INITIAL bead-step count
// (docs/bead-lattice.md "The count") and straight-segment endpoints.
//   - destWire: "destNode.destPort" → *PacedWire (owned by the destination).
//   - edgeWire: edge label → *PacedWire (same pointer; for stdin_reader lookup).
//   - edgeEndpoints: edge label → source/target node IDs + handles (for NodeMoveRegistry).
//   - edgeSteps: each edge's OWN bead-step count (per-edge geometry).
//   - edgeSegments: each edge's straight-segment endpoints (Start/End) so the
//     bead's position stream evaluates P(t)=Start+t*(End-Start).
//
// All keyed by edge label; consumed by buildNodes when binding the source Out.
func (b *buildCtx) allocateWires() {
	destWire := map[string]*wire.PacedWire{}
	edgeWire := WireRegistry{}
	edgeEndpoints := map[string]EdgeEndpoints{}
	edgeSteps := map[string]int{}
	edgeSegments := map[string]wireSegment{}
	for _, e := range b.spec.Edges {
		destKey := e.Target + "." + e.TargetHandle
		// Segment: node SURFACE to node surface (docs/channels-not-ports.md), the GPU
		// boundary the renderer draws from (edgeSegment) — no port name feeds this any
		// more, a port contributes no geometry.
		srcG, tgtG := b.nodeGeoms[e.Source], b.nodeGeoms[e.Target]
		seg := edgeSegment(srcG, tgtG)
		// Steps: the LIVE center-to-center distance between the two nodes, run through
		// edgeStepCount (docs/bead-lattice.md "The count") — the SAME function and the
		// SAME kind of distance the source node's own chainBeads pass (chain_beads.go)
		// will keep recomputing once running, so this load-time value and that first
		// live pass can never disagree.
		dist := nodeWorldPos(srcG).Sub(nodeWorldPos(tgtG)).Length()
		steps := edgeStepCount(dist, srcG.Kind, tgtG.Kind)
		edgeSteps[e.Label] = steps
		edgeSegments[e.Label] = seg
		// One wire per destination input port, and — since fan-in is removed
		// (validateNoFanIn) — exactly one edge per port, so this is strictly one wire per
		// edge. A destKey already present means a fan-in spec slipped past the parser: that
		// is a build-invariant violation, not a topology to silently share a wire for.
		if _, exists := destWire[destKey]; exists {
			panic("allocateWires: two edges target " + destKey + " — validateNoFanIn should have rejected this fan-in at parse")
		}
		// wire.DwellTicksPerBead is the ONE canonical dwell-per-step constant
		// (docs/bead-lattice.md "Timing" — uniform pulse speed is now structural,
		// not a length divided by a speed); guarded as the sole non-test
		// NewPacedWire call site by tools/check-uniform-pulse-speed.sh.
		pw := wire.NewPacedWire(steps, wire.DwellTicksPerBead)
		pw.Target = e.Target
		pw.TargetHandle = e.TargetHandle
		pw.Trace = b.tr
		destWire[destKey] = pw
		edgeWire[e.Label] = pw
		edgeEndpoints[e.Label] = EdgeEndpoints{
			Source: e.Source, Target: e.Target,
			SourceHandle: e.SourceHandle, TargetHandle: e.TargetHandle,
		}
	}
	b.destWire = destWire
	b.edgeWire = edgeWire
	b.edgeEndpoints = edgeEndpoints
	b.edgeSteps = edgeSteps
	b.edgeSegments = edgeSegments
}

// buildMoveDispatch builds the MoveDispatch from initial geometry and edge
// endpoints. It creates one nodeMover per node and one edgeMover per edge; each
// owns its geometry and recomputes itself on a node-move (no central
// coordinator). The trace lets each mover stream its own node/edge geometry on a
// move. Outs + dest wires are bound later (bindDispatch) once node construction
// has populated them. Also declares the double-link movement graph (links.go;
// polar locks ride on it in a later step — the lock system and the central polar
// position store have been removed, so node positions live in the movers' held
// geometry) and installs the aimed-port registry for drag-time aiming.
func (b *buildCtx) buildMoveDispatch() error {
	// SPEC order (b.spec.Nodes/Edges — the deterministic directory-sorted order parseSpec
	// read the topology in), NOT map iteration order, so the buffer's row seed
	// (md.NodeSeeds/EdgeSeeds) gives every node/edge a deterministic row.
	nodeOrder := make([]string, len(b.spec.Nodes))
	for i, n := range b.spec.Nodes {
		nodeOrder[i] = n.ID
	}
	edgeOrder := make([]string, len(b.spec.Edges))
	for i, e := range b.spec.Edges {
		edgeOrder[i] = e.Label
	}
	md, err := newMoveDispatch(b.nodeGeoms, b.edgeEndpoints, b.tr, nodeOrder, edgeOrder, b.clk, &b.speedSinks, b.spec.RowCount)
	if err != nil {
		return fmt.Errorf("buildMoveDispatch: %w", err)
	}
	if b.hasScene {
		// Persisted scene sphere: install it now so md.ui.sceneSphere is consistent straight out
		// of LoadTopology (a fresh/legacy scene has none — main.go's LoadSceneSphere then
		// content-fits it from the loaded node centers).
		md.ui.sceneSphere = b.sphere
	}

	// The quantized layout is authoritative by default — b.quantizedOffsets was already
	// resolved (stored offset, or measured from the pre-quantized center) by
	// computeQuantizedLayout, which also overwrote b.nodeGeoms so the nodeMovers newMoveDispatch
	// just built above are already seeded from the composed centers. Seed each node's OWN
	// mover field (nodeMover.quantOffset) from it here — there is no shared md.quantizedOffsets
	// map anymore (that map, read/written by multiple mover goroutines for different keys,
	// was the "concurrent map read and map write" fatal fixed by node6-drag-decentralized.md's
	// per-node ownership). A node missing an entry in b.quantizedOffsets keeps its
	// nodeMover's zero-value quantOffset, matching the old map's zero-value-on-miss read.
	// PER SCENE, not always-on (scene_tabs.go's QuantizedDrag): a bead-distance step is
	// invisible in a scene that is large against it and dominant in one that is not.
	md.lq.quantizedLayout = SceneUsesQuantizedDrag(b.scenePath)
	// Per scene as well (scene_tabs.go's CoplanarEdges): each node's own copy, set here on
	// the single-threaded build path, read afterwards only by that node's own goroutine.
	if SceneWantsCoplanarEdges(b.scenePath) {
		for _, nm := range md.mr.nodeMovers {
			nm.coplanarEdges = true
		}
	}
	if SceneWantsUpAxis(b.scenePath) {
		for _, nm := range md.mr.nodeMovers {
			nm.upAxis = true
		}
	}
	for id, off := range b.quantizedOffsets {
		if nm, ok := md.mr.nodeMovers[id]; ok {
			nm.quantOffset = off
		}
	}
	// Seed each node's OWN selfKind (specNode.Type), set once at construction.
	for _, n := range b.spec.Nodes {
		nm, ok := md.mr.nodeMovers[n.ID]
		if !ok {
			continue
		}
		nm.selfKind = n.Type
		if n.TiltVectorThetaIdx != nil {
			nm.tiltVectorThetaIdx = *n.TiltVectorThetaIdx
		}
		if n.TiltVectorPhiIdx != nil {
			nm.tiltVectorPhiIdx = *n.TiltVectorPhiIdx
		}
	}
	// Seed each node's OWN neighborKinds map — every DIRECT domain-adjacent neighbor id
	// mapped to that neighbor's own kind name, derived straight from the loaded spec's
	// node list + edges (no separate persisted file: adjacency is already known from
	// b.spec.Edges, and a node's kind is already known from b.spec.Nodes, so keeping a
	// second stored copy in sync would only be a second place for the two to drift).
	// UNDIRECTED: both endpoints of every edge learn the other's kind.
	kindByID := make(map[string]string, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		kindByID[n.ID] = n.Type
	}
	linkNeighborKind := func(fromID, toID string) {
		nm, ok := md.mr.nodeMovers[fromID]
		if !ok {
			return
		}
		if nm.neighborKinds == nil {
			nm.neighborKinds = map[string]string{}
		}
		nm.neighborKinds[toID] = kindByID[toID]
	}
	for _, e := range b.spec.Edges {
		linkNeighborKind(e.Source, e.Target)
		linkNeighborKind(e.Target, e.Source)
	}
	// Seed each node's OWN outgoing-edge targets (nodeMover.outTargets) — the chains it
	// owns (chain_beads.go). A chain belongs to exactly one endpoint: the source, matching
	// where the edge is stored on disk.
	for _, e := range b.spec.Edges {
		nm, ok := md.mr.nodeMovers[e.Source]
		if !ok {
			continue
		}
		nm.outTargets = append(nm.outTargets, e.Target)
	}
	b.md = md
	return nil
}

// buildTypeMaps builds the id→type map and per-kind Broadcast port set (needed
// for sourceHandle normalization in buildEdgeMaps).
func (b *buildCtx) buildTypeMaps() {
	nodeType := map[string]string{}
	for _, n := range b.spec.Nodes {
		nodeType[n.ID] = n.Type
	}
	kindBroadcastPorts := map[string]map[string]bool{}
	for kind, bind := range Registry {
		outMultis := map[string]bool{}
		for _, p := range bind.Ports {
			if p.Dir == PortBroadcast {
				outMultis[p.Name] = true
			}
		}
		kindBroadcastPorts[kind] = outMultis
	}
	b.nodeType = nodeType
	b.kindBroadcastPorts = kindBroadcastPorts
}

// buildEdgeMaps builds the inbound and outbound edge maps.
//   - inbound:  target node id → port name → destKey ("destNode.destPort")
//   - outbound: source node id → port name → []edge label
//   - outboundHandle: source node id → port name → []sourceHandle (indexed, same order as outbound)
//
// For Broadcast ports, sourceHandle may be "<portName><index>" — normalize to portName.
func (b *buildCtx) buildEdgeMaps() {
	inbound := map[string]map[string]string{}
	outbound := map[string]map[string][]string{}
	outboundHandle := map[string]map[string][]string{}
	for _, e := range b.spec.Edges {
		if inbound[e.Target] == nil {
			inbound[e.Target] = map[string]string{}
		}
		if outbound[e.Source] == nil {
			outbound[e.Source] = map[string][]string{}
		}
		if outboundHandle[e.Source] == nil {
			outboundHandle[e.Source] = map[string][]string{}
		}
		inbound[e.Target][e.TargetHandle] = e.Target + "." + e.TargetHandle
		srcKey := e.SourceHandle
		if base, isMulti := broadcastBaseName(e.SourceHandle, b.nodeType[e.Source], b.kindBroadcastPorts); isMulti {
			srcKey = base
		}
		outbound[e.Source][srcKey] = append(outbound[e.Source][srcKey], e.Label)
		outboundHandle[e.Source][srcKey] = append(outboundHandle[e.Source][srcKey], e.SourceHandle)
	}
	b.inbound = inbound
	b.outbound = outbound
	b.outboundHandle = outboundHandle
}

// buildNodes builds each node from the wire allocation and edge maps computed by
// earlier phases. outSink collects every paced source Out keyed by "node.handle"
// so node-move can update per-edge travel-time on the Out.
func (b *buildCtx) buildNodes() error {
	outSink := map[string]*wire.Out{}
	nodes := make([]wire.Node, 0, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		bind := Registry[n.Type]
		pb := newPortBindings()
		pb.outSink = outSink
		pb.clock = b.clk // shared clock for clock-paced interior animation (Input refill slide)
		// &b.speedSinks (not a fresh slice per node): every node's channels append
		// onto the SAME build-wide accumulator, so LoadTopology's one returned
		// list carries every clock-owning goroutine across the whole build.
		pb.speedSinks = &b.speedSinks
		// md gives injectClosures's interior-bead Emit* closures access to this node's
		// OWN dedicated interior fd (md.sw.interiorOuts, keyed by node id) + the injected
		// frame builder (md.sw.buildInteriorFrame) — the SECOND emitting goroutine per node
		// (memory/feedback_no_single_writer_bridge.md). nil until SetNodeStreams runs
		// (main.go, after LoadTopology returns); the Emit* closures nil-check both before
		// writing and no-op until then.
		pb.md = b.md

		for _, port := range bind.Ports {
			switch port.Dir {
			case PortIn:
				dk, ok := b.inbound[n.ID][port.Name]
				if ok {
					pb.SetSinglePaced(port.Name, b.destWire[dk])
				}
				// If no inbound edge, a.In() falls back to a dead-end chan.

			case PortOut:
				labels := b.outbound[n.ID][port.Name]
				if len(labels) > 0 {
					// Look up wire by destination of the first outbound edge.
					// For fan-in, the destination port owns the wire.
					// Send rule is node-owned, keyed by this output port name.
					rule := nodeSendRule(n, port.Name)
					lbl := labels[0]
					pb.SetSinglePacedRule(port.Name, b.edgeWire[lbl], rule, b.edgeSteps[lbl], b.edgeSegments[lbl], lbl)
				}
				// If no outbound edge, a.Out() falls back to a dead-end chan.

			case PortBroadcast:
				labels := b.outbound[n.ID][port.Name]
				handles := b.outboundHandle[n.ID][port.Name]
				for i, lbl := range labels {
					handle := port.Name
					if i < len(handles) {
						handle = handles[i]
					}
					// Per-port (per fan-out element): the rule is keyed by the
					// concrete output port name (sourceHandle, e.g. "ToNext0").
					rule := nodeSendRule(n, handle)
					pb.AppendBroadcastWithHandle(port.Name, handle, b.edgeWire[lbl], rule, b.edgeSteps[lbl], b.edgeSegments[lbl], lbl)
				}
				// If no outbound edges, builder falls back to a dead-end slice.
			}
		}

		var tiltThetaIdx, tiltPhiIdx int32
		if n.TiltVectorThetaIdx != nil {
			tiltThetaIdx = *n.TiltVectorThetaIdx
		}
		if n.TiltVectorPhiIdx != nil {
			tiltPhiIdx = *n.TiltVectorPhiIdx
		}
		nd, err := bind.Build(b.ctx, n.ID, n.Data, pb, b.tr, b.nodeGeoms[n.ID], tiltThetaIdx, tiltPhiIdx)
		if err != nil {
			return fmt.Errorf("LoadTopology: build node %q: %w", n.ID, err)
		}
		nodes = append(nodes, nd)
	}
	b.outSink = outSink
	b.nodes = nodes
	return nil
}

// bindDispatch binds per-edge source Outs and dest wires into each edgeMover so
// a node-move updates per-edge travel-time.
func (b *buildCtx) bindDispatch() {
	b.md.Bind(b.outSink, SlotRegistry(b.destWire))
}
