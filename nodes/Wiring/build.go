// build.go — turns an already-parsed and validated topoSpec into the running
// graph: node geometry, quantized layout, wire allocation, the MoveDispatch,
// and the built []wire.Node themselves. buildFromSpec orchestrates the phase
// helpers below (all methods on buildCtx) in the same order the original
// monolithic loader.go function performed them; behavior is unchanged.
// loader.go's LoadTopology/LoadTopologyFromJSON call buildFromSpec after
// parsing + validating via topo_spec.go.

package Wiring

import (
	"context"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"reflect"

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

	// Phase 1: node geometry + world centers.
	nodeGeoms map[string]nodeGeom
	centers   map[string]vec3

	// Phase 1b: quantized flat absolute scene-polar layout (quantized_layout.go) —
	// resolved BEFORE reach/wire/dispatch phases so every later phase computes from the
	// COMPOSED (authoritative) centers, not the raw loaded ones. Every node is a root
	// measured about the scene center — no reference/parent concept.
	quantizedOffsets map[string]quantizedOffset

	// Phase 1c: double-link LOCAL POLAR data (layout_holder.go) — every domain
	// double-link (bidirectional edge pair) gives each endpoint its own local
	// polar to the other, measured with ITSELF as center. Computed AFTER the
	// quantized layout so it reads the composed (authoritative) centers, and
	// injected into each built node's LocalPolars field (buildNodes) — additive,
	// does not feed back into position (quantizedOffsets stays authoritative).
	localPolars map[string][]wire.LocalPolar

	// localPoles is the per-node measurement pole (rotating_pole.go localPole result)
	// that localPolars[id]'s entries were quantized about — either the stored
	// (specNode.LocalPoleTheta/Phi) value carried forward, or freshly computed from live
	// composed centers when no stored pole exists. Attached to each node's LayoutHolder
	// (buildNodes, via LayoutHolder.SetPole) so a later drag's requantizePoleTraced
	// reconstructs unchanged neighbors against the SAME pole this load quantized about.
	localPoles map[string]dir

	// Phase 4: per-destination-port wire allocation + per-edge geometry.
	destWire      map[string]*wire.PacedWire
	edgeWire      WireRegistry
	edgeEndpoints map[string]EdgeEndpoints
	edgeArc       map[string]float64
	edgeLatency   map[string]float64
	edgeSegments  map[string]wireSegment

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
func buildFromSpec(ctx context.Context, spec topoSpec, tr *T.Trace, clk wire.Clock, sphere sceneSphere, hasScene bool) ([]wire.Node, SlotRegistry, *MoveDispatch, []chan float64, error) {
	b := &buildCtx{ctx: ctx, spec: spec, tr: tr, clk: clk, sphere: sphere, hasScene: hasScene}

	b.computeNodeGeometry()
	b.computeQuantizedLayout()
	b.computeLocalPolars()
	b.computeReachRadii()
	b.allocateWires()
	b.buildMoveDispatch()
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
// fan-in is rejected at parse) and
// computes each edge's own travel-time (arc length / sim latency) and
// straight-segment endpoints.
//   - destWire: "destNode.destPort" → *PacedWire (owned by the destination).
//   - edgeWire: edge label → *PacedWire (same pointer; for stdin_reader lookup).
//   - edgeEndpoints: edge label → source/target node IDs + handles (for NodeMoveRegistry).
//   - edgeArc / edgeLatency: each edge's OWN travel-time (per-edge geometry).
//   - edgeSegments: each edge's straight-segment endpoints (Start/End) so the
//     bead's position stream evaluates P(t)=Start+t*(End-Start).
//
// All keyed by edge label; consumed by buildNodes when binding the source Out.
func (b *buildCtx) allocateWires() {
	destWire := map[string]*wire.PacedWire{}
	edgeWire := WireRegistry{}
	edgeEndpoints := map[string]EdgeEndpoints{}
	edgeArc := map[string]float64{}
	edgeLatency := map[string]float64{}
	edgeSegments := map[string]wireSegment{}
	for _, e := range b.spec.Edges {
		destKey := e.Target + "." + e.TargetHandle
		// Per-edge segment + arc, node-to-node (polar-frame-rewrite.md option A). The arc
		// (pulse travel budget) is the polar law-of-cosines distance between the two node
		// positions (edgeArcPolar) — pure polar. The segment is the world node-to-node line
		// for the renderer (edgeSegment), the GPU-boundary cartesian.
		srcG, tgtG := b.nodeGeoms[e.Source], b.nodeGeoms[e.Target]
		seg := edgeSegment(srcG, tgtG, e.SourceHandle, e.TargetHandle)
		arcLength := edgeArcPolar(srcG, tgtG, e.SourceHandle, e.TargetHandle)
		simLatencyMs := arcLength / wire.PulseSpeedWuPerMs
		edgeArc[e.Label] = arcLength
		edgeLatency[e.Label] = simLatencyMs
		edgeSegments[e.Label] = seg
		// One wire per destination input port, and — since fan-in is removed
		// (validateNoFanIn) — exactly one edge per port, so this is strictly one wire per
		// edge. A destKey already present means a fan-in spec slipped past the parser: that
		// is a build-invariant violation, not a topology to silently share a wire for.
		if _, exists := destWire[destKey]; exists {
			panic("allocateWires: two edges target " + destKey + " — validateNoFanIn should have rejected this fan-in at parse")
		}
		pw := wire.NewPacedWire(arcLength, wire.PulseSpeedWuPerTick)
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
	b.edgeArc = edgeArc
	b.edgeLatency = edgeLatency
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
func (b *buildCtx) buildMoveDispatch() {
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
	md := newMoveDispatch(b.nodeGeoms, b.edgeEndpoints, b.tr, nodeOrder, edgeOrder, b.clk, &b.speedSinks)
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
	md.lq.quantizedLayout = true
	for id, off := range b.quantizedOffsets {
		if nm, ok := md.mr.nodeMovers[id]; ok {
			nm.quantOffset = off
		}
	}
	// Seed each node's OWN cascadeEdges (nodeMover.cascadeEdges doc comment) from the
	// STORED per-node cascade-edges.json file (specNode.CascadeEdges, loader_tree.go) —
	// this is now the SOLE source of truth for both delta-forward propagation
	// (nodeMover.forwardDelta) and the cascade-link overlay's rendered pairs. It is
	// hand-authored/persisted data, not derived from the local-polar adjacency at load
	// (the computed cascade_links.go machinery was removed). Absent file → empty list
	// (readJSONBestEffort's missing-file default), matching every other per-node
	// optional-file convention in this loader.
	for _, n := range b.spec.Nodes {
		nm, ok := md.mr.nodeMovers[n.ID]
		if !ok {
			continue
		}
		nm.cascadeEdges = n.CascadeEdges
		nm.selfKind = n.Type
		nm.cascadeKinds = n.CascadeKinds
	}
	b.md = md
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
					pb.SetSinglePacedRule(port.Name, b.edgeWire[lbl], rule, b.edgeArc[lbl], b.edgeLatency[lbl], b.edgeSegments[lbl], lbl)
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
					pb.AppendBroadcastWithHandle(port.Name, handle, b.edgeWire[lbl], rule, b.edgeArc[lbl], b.edgeLatency[lbl], b.edgeSegments[lbl], lbl)
				}
				// If no outbound edges, builder falls back to a dead-end slice.
			}
		}

		// Reuse the exact partnerCenter lookup already installed on this node's mover
		// (buildMoveDispatch runs before buildNodes) so the INITIAL geometry emit and every
		// later re-emit compute a connected port's aim identically.
		var pc partnerCenterFn
		if nm, ok := b.md.mr.nodeMovers[n.ID]; ok {
			pc = nm.partnerCenter
		}
		nd, err := bind.Build(b.ctx, n.ID, n.Data, pb, b.tr, b.nodeGeoms[n.ID], pc)
		if err != nil {
			return fmt.Errorf("LoadTopology: build node %q: %w", n.ID, err)
		}
		// Attach this node's computed LocalPolars list (layout_holder.go) via the
		// promoted embedded Wiring.LayoutHolder every kind gets — so the node's
		// layout goroutine owns it without per-kind wiring. Locate the embedded
		// *LayoutHolder by reflection (same field-lookup builders.go/loader.go use
		// elsewhere for port/data injection), then load through its own setter
		// rather than reflecting on the unexported localPolars field directly, so
		// this initial load goes through the same entry point every other
		// LocalPolars access does.
		if v := reflect.ValueOf(nd).Elem(); v.Kind() == reflect.Struct {
			if lhField := v.FieldByName("LayoutHolder"); lhField.IsValid() && lhField.CanAddr() {
				if lh, ok := lhField.Addr().Interface().(*wire.LayoutHolder); ok {
					if lps, ok := b.localPolars[n.ID]; ok {
						lh.LoadLocalPolars(lps)
					}
					// Attach the measurement pole this load quantized LocalPolars about
					// (computeLocalPolars: stored pole honored verbatim, else resolved
					// fresh from live geometry) so a LATER drag's requantizePoleTraced
					// reconstructs unchanged neighbors against the SAME pole, not an
					// assumed home pole.
					if pole, ok := b.localPoles[n.ID]; ok {
						lh.SetPole(pole)
					}
					// Register this node's embedded *Wiring.LayoutHolder with the move
					// dispatcher so a later drag (RootMove) can route a local-polar
					// re-quantize to the OWNING node's own holder — MoveDispatch never
					// copies or owns LocalPolars itself.
					b.md.lq.layoutHolders[n.ID] = lh
				}
			}
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
