// node_geometry_wire.go — nodeGeometry's own single-threaded WIRING API: the methods the
// loader's construction pass (move_dispatch_construct.go, build_move_dispatch.go,
// mover_registry.go's bind, stream_wiring.go, move_streams.go, move_persist.go,
// build_args_selfdrive.go) calls to seed a node's sub-owners, instead of reaching into
// nine different unexported sub-structs by field (docs/planning/movedispatch-decomposition.md
// §19). Every method here runs during the single-threaded wiring pass, before any driving
// goroutine starts — none of them is called again once a node is running. This is NOT the
// §18 accessor/actor-move surface: these stay unexported, package-Wiring-only, and exist
// purely so the loader assigns through a named call instead of a bare field write.
package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// wireMessaging installs this node's whole routing surface in one call: how it resolves
// another id to a send func, how it hands a move off to its own retry queue, the test-only
// tap, how it resolves a neighbor's center, and its owner-goroutine commit path. All five
// are set exactly once, together, right after newNodeGeometry — see
// move_dispatch_construct.go's per-node loop, the only call site.
func (m *nodeGeometry) wireMessaging(
	resolveDest func(id string) (func(movemsg.Msg) bool, bool),
	sendMove func(id string, msg movemsg.Msg),
	tap func(destID string, msg movemsg.Msg),
	centerOf func(id string) (vec3, bool),
	commitLocal func(id string, newPos vec3),
) {
	m.msg.resolveDest = resolveDest
	m.msg.sendMove = sendMove
	m.msg.tap = tap
	m.msg.centerOf = centerOf
	m.msg.commitLocal = commitLocal
}

// setMsgTap installs (or clears, with nil) the test-only message-trace hook on this node
// alone — MoveDispatch.SetMsgTap (move_streams.go) loops every nodeMover and calls this once
// per node.
func (m *nodeGeometry) setMsgTap(tap func(destID string, msg movemsg.Msg)) {
	m.msg.tap = tap
}

// ensureNeighborChannel makes this node's dedicated inbound channel for otherID if it does
// not already exist — the "two channels, A→B and B→A" per-pair topology, one direction at a
// time (move_dispatch_construct.go calls it once per endpoint of every edge).
func (m *nodeGeometry) ensureNeighborChannel(otherID string) {
	if _, exists := m.msg.neighborIn[otherID]; !exists {
		m.msg.neighborIn[otherID] = make(chan movemsg.Msg, moverInboxDepth)
	}
}

// addMutualTarget records that target is one of this node's own mutual (points-back-at-me)
// targets, lazily allocating the map on first use.
func (m *nodeGeometry) addMutualTarget(target string) {
	if m.topo.mutualTargets == nil {
		m.topo.mutualTargets = map[string]bool{}
	}
	m.topo.mutualTargets[target] = true
}

// seedPartnerCenter seeds this node's own copy of a direct neighbor's last-known world
// center — the load-time seed move_dispatch_construct.go performs before md.Start, later
// kept current by that neighbor's own movemsg.KindNeighborCenter push (handle, above).
func (m *nodeGeometry) seedPartnerCenter(neighborID string, c vec3) {
	if m.topo.partnerCenters == nil {
		m.topo.partnerCenters = map[string]vec3{}
	}
	m.topo.partnerCenters[neighborID] = c
}

// addEdgeID appends edgeID to this node's own list of incident edges.
func (m *nodeGeometry) addEdgeID(edgeID string) {
	m.topo.edgeIDs = append(m.topo.edgeIDs, edgeID)
}

// addNeighborKind records toID's kind in this node's own neighbor-kind map, lazily
// allocating on first use — build_move_dispatch.go's linkNeighborKind closure calls this
// once per directed adjacency.
func (m *nodeGeometry) addNeighborKind(toID, kind string) {
	if m.topo.neighborKinds == nil {
		m.topo.neighborKinds = map[string]string{}
	}
	m.topo.neighborKinds[toID] = kind
}

// setSceneFlags installs this node's two scene-wide ring-axis drawing choices in one call —
// build_move_dispatch.go's two separate coplanar/up-axis loops become one loop over every
// node, each calling this once.
func (m *nodeGeometry) setSceneFlags(coplanarEdges, upAxis bool) {
	m.flags.coplanarEdges = coplanarEdges
	m.flags.upAxis = upAxis
}

// setQuantOffset installs this node's initial quantized polar offset at CONSTRUCTION time
// (build_move_dispatch.go, seeded from the loader's computed/persisted offset). This is
// distinct from the RUNTIME write in commit_node_move.go (nm.quantOffset = off on this
// node's own driving goroutine, mutating it on every commit) — that call site stays a bare
// field write on purpose, unchanged by this method's existence.
func (m *nodeGeometry) setQuantOffset(off quantoffset.QuantizedOffset) {
	m.quantOffset = off
}

// setTopTiltVectorThetaIdx seeds this node's own top-tilt-vector lattice index from the
// spec, at construction.
func (m *nodeGeometry) setTopTiltVectorThetaIdx(idx int32) {
	m.tilt.topTiltVectorThetaIdx = idx
}

// addOutTarget appends target to this node's own outgoing chain targets.
func (m *nodeGeometry) addOutTarget(target string) {
	m.outs.outTargets = append(m.outs.outTargets, target)
}

// addOutWire bundles the four index-parallel nodeOuts appends moverRegistry.bind makes for
// one edge leaving this node: the wire itself, its target id, the source Out this edge's
// step count publishes through (may be nil — see nodeOuts.outWireOuts' own doc comment),
// and the matching edgeMover's own SendSteps bound method.
func (m *nodeGeometry) addOutWire(pw *wire.PacedWire, target string, o *wire.Out, sendSteps func(int)) {
	m.outs.outWires = append(m.outs.outWires, pw)
	m.outs.outWireTargets = append(m.outs.outWireTargets, target)
	m.outs.outWireOuts = append(m.outs.outWireOuts, o)
	m.outs.outStepsIn = append(m.outs.outStepsIn, sendSteps)
}

// wireStream installs this node's dedicated content-buffer stream: its claimed fd, its
// stable buffer row, its static kind column, the neighbor-row lookup its own events need,
// and its frame packer — everything setNodeStreams seeds in one pass, one node at a time.
func (m *nodeGeometry) wireStream(streamOut claimedStream, row int32, kindID uint8, nodeRowFor func(id string) (int32, bool), buildFrame NodeFrameBuilder) {
	m.stream.streamOut = streamOut
	m.stream.nodeRow = row
	m.stream.kindID = kindID
	m.topo.nodeRowFor = nodeRowFor
	m.stream.buildFrame = buildFrame
}

// setPersistRoot installs the tree root this node writes its own per-node files into — set
// once, for every node, by MoveDispatch.EnableEditPersist after the startup seed.
func (m *nodeGeometry) setPersistRoot(root string) {
	m.persistRoot = root
}

// copyClockSrc copies this node's clock source into its own live clock, exactly once. Used
// by ClaimSelfDrive (build_args_selfdrive.go) for a self-driven node, which has no
// nodeMover.run goroutine-start to perform this copy for it; a ring node's own
// nodeMover.run (node_mover.go, part of the actor's own files) performs the identical copy
// inline at its own goroutine start, not through this method — that call site is
// POST-construction (it runs on the node's own goroutine, not the single-threaded wiring
// pass) and is left as-is.
func (m *nodeGeometry) copyClockSrc() {
	if m.clocks.clockSrc != nil {
		m.clocks.clk = m.clocks.clockSrc.Copy()
	}
}
