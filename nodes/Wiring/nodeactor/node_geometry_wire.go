// node_geometry_wire.go — NodeGeometry's own single-threaded WIRING API: the exported
// methods package Wiring's construction pass (move_dispatch_construct.go,
// build_move_dispatch.go, mover_registry.go's bind, stream_wiring.go, move_streams.go,
// move_persist.go, build_args_selfdrive.go) calls to seed a node's sub-owners, instead of
// reaching into nine different unexported sub-structs by field
// (docs/planning/movedispatch-decomposition.md §19). Every method here runs during the
// single-threaded wiring pass, before any driving goroutine starts — none of them is
// called again once a node is running.
//
// §20 exported this whole file's method set (previously package-Wiring-only, §19's own
// "NOT the §18 accessor/actor-move surface" note) because the type itself left package
// Wiring in §20 — package Wiring is now an external caller across a real package
// boundary, so these calls MUST cross it through named methods rather than field writes.
// No method here reaches a channel directly by returning it — see
// node_geometry_accessors.go for the channel-wrapping half of this exported surface
// (NeighborTrySend/SendExternal/EnqueueSend/PollCenter), which follows §17's edgeMover
// rule: no channel is ever exported, only a method that closes over one.
package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// WireMessaging installs this node's whole routing surface in one call: how it resolves
// another id to a send func, how it hands a move off to its own retry queue, the test-only
// tap, how it resolves a neighbor's center, and its owner-goroutine commit path. All five
// are set exactly once, together, right after NewNodeGeometry — see package Wiring's
// move_dispatch_construct.go per-node loop, the only call site.
func (m *NodeGeometry) WireMessaging(
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

// SetMsgTap installs (or clears, with nil) the test-only message-trace hook on this node
// alone — package Wiring's MoveDispatch.SetMsgTap loops every NodeMover's own geometry and
// calls this once per node.
func (m *NodeGeometry) SetMsgTap(tap func(destID string, msg movemsg.Msg)) {
	m.msg.tap = tap
}

// EnsureNeighborChannel makes this node's dedicated inbound channel for otherID if it does
// not already exist — the "two channels, A→B and B→A" per-pair topology, one direction at a
// time (package Wiring's move_dispatch_construct.go calls it once per endpoint of every
// edge).
func (m *NodeGeometry) EnsureNeighborChannel(otherID string) {
	if _, exists := m.msg.neighborIn[otherID]; !exists {
		m.msg.neighborIn[otherID] = make(chan movemsg.Msg, inboxDepth)
	}
}

// AddMutualTarget records that target is one of this node's own mutual (points-back-at-me)
// targets, lazily allocating the map on first use.
func (m *NodeGeometry) AddMutualTarget(target string) {
	if m.topo.mutualTargets == nil {
		m.topo.mutualTargets = map[string]bool{}
	}
	m.topo.mutualTargets[target] = true
}

// SeedPartnerCenter seeds this node's own copy of a direct neighbor's last-known world
// center — the load-time seed package Wiring's move_dispatch_construct.go performs before
// md.Start, later kept current by that neighbor's own movemsg.KindNeighborCenter push
// (handle, node_geometry.go).
func (m *NodeGeometry) SeedPartnerCenter(neighborID string, c vec3) {
	if m.topo.partnerCenters == nil {
		m.topo.partnerCenters = map[string]vec3{}
	}
	m.topo.partnerCenters[neighborID] = c
}

// AddEdgeID appends edgeID to this node's own list of incident edges.
func (m *NodeGeometry) AddEdgeID(edgeID string) {
	m.topo.edgeIDs = append(m.topo.edgeIDs, edgeID)
}

// AddNeighborKind records toID's kind in this node's own neighbor-kind map, lazily
// allocating on first use — package Wiring's build_move_dispatch.go linkNeighborKind
// closure calls this once per directed adjacency.
func (m *NodeGeometry) AddNeighborKind(toID, kind string) {
	if m.topo.neighborKinds == nil {
		m.topo.neighborKinds = map[string]string{}
	}
	m.topo.neighborKinds[toID] = kind
}

// SetSelfKind installs this node's own kind name at construction — package Wiring's
// build_move_dispatch.go seeds it once, per node, from the loaded spec (specNode.Type).
// Added in §20 (movedispatch-decomposition.md), replacing what was a bare field write
// (`nm.selfKind = n.Type`) while this type still lived in package Wiring; §19 left that
// write as a bare field on purpose since selfKind was not one of the 9 sub-structs it
// scoped, but the package move makes any external field write unreachable, so it is
// absorbed here like every other construction-time write in this file.
func (m *NodeGeometry) SetSelfKind(kind string) {
	m.selfKind = kind
}

// SetSceneFlags installs this node's two scene-wide ring-axis drawing choices in one call —
// package Wiring's build_move_dispatch.go two separate coplanar/up-axis loops become one
// loop over every node, each calling this once.
func (m *NodeGeometry) SetSceneFlags(coplanarEdges, upAxis bool) {
	m.flags.coplanarEdges = coplanarEdges
	m.flags.upAxis = upAxis
}

// SetQuantOffset installs this node's initial quantized polar offset at CONSTRUCTION time
// (package Wiring's build_move_dispatch.go, seeded from the loader's computed/persisted
// offset). This is distinct from the RUNTIME write CommitQuantOffset makes (on this node's
// own driving goroutine, mutating it on every commit) — see node_geometry_accessors.go.
func (m *NodeGeometry) SetQuantOffset(off quantoffset.QuantizedOffset) {
	m.quantOffset = off
}

// SetTopTiltVectorThetaIdx seeds this node's own top-tilt-vector lattice index from the
// spec, at construction.
func (m *NodeGeometry) SetTopTiltVectorThetaIdx(idx int32) {
	m.tilt.topTiltVectorThetaIdx = idx
}

// AddOutTarget appends target to this node's own outgoing chain targets.
func (m *NodeGeometry) AddOutTarget(target string) {
	m.outs.outTargets = append(m.outs.outTargets, target)
}

// AddOutWire bundles the four index-parallel nodeOuts appends package Wiring's
// moverRegistry.bind makes for one edge leaving this node: the wire itself, its target id,
// the source Out this edge's step count publishes through (may be nil — see
// nodeOuts.outWireOuts' own doc comment), and the matching edgeMover's own SendSteps bound
// method.
func (m *NodeGeometry) AddOutWire(pw *wire.PacedWire, target string, o *wire.Out, sendSteps func(int)) {
	m.outs.outWires = append(m.outs.outWires, pw)
	m.outs.outWireTargets = append(m.outs.outWireTargets, target)
	m.outs.outWireOuts = append(m.outs.outWireOuts, o)
	m.outs.outStepsIn = append(m.outs.outStepsIn, sendSteps)
}

// WireStream installs this node's dedicated content-buffer stream: its claimed fd, its
// stable buffer row, its static kind column, the neighbor-row lookup its own events need,
// and its frame packer — everything package Wiring's setNodeStreams seeds in one pass, one
// node at a time.
//
// streamOut is a nodeactor.StreamHandle (stream_claim.go), this package's OWN duplicate of
// package Wiring's unexported claimedStream (same duplication precedent as
// nodes/Wiring/edgemover's own stream_claim.go, §17) — claimed against this package's OWN
// ClaimRegistry, never package Wiring's. This is safe because node claim keys (node ids)
// and edge claim keys (edge labels) and the VIEW claim key (a fixed singleton, in
// viewstate's own separate registry) can never collide: package Wiring's own three claim
// registries were already namespaced this way BEFORE any of them left the package (see
// package Wiring's stream_wiring.go doc comment — "the three can never collide (disjoint
// namespaces...)"), so splitting node claims into their own registry, in their own
// package, changes nothing observable.
func (m *NodeGeometry) WireStream(streamOut StreamHandle, row int32, kindID uint8, nodeRowFor func(id string) (int32, bool), buildFrame NodeFrameBuilder) {
	m.stream.streamOut = streamOut
	m.stream.nodeRow = row
	m.stream.kindID = kindID
	m.topo.nodeRowFor = nodeRowFor
	m.stream.buildFrame = buildFrame
}

// SetPersistRoot installs the tree root this node writes its own per-node files into — set
// once, for every node, by package Wiring's MoveDispatch.EnableEditPersist after the
// startup seed.
func (m *NodeGeometry) SetPersistRoot(root string) {
	m.persistRoot = root
}

// CopyClockSrc copies this node's clock source into its own live clock, exactly once. Used
// by package Wiring's ClaimSelfDrive (build_args_selfdrive.go) for a self-driven node,
// which has no NodeMover.Run goroutine-start to perform this copy for it; a ring node's
// own NodeMover.Run (node_mover.go, part of the actor's own files) performs the identical
// copy inline at its own goroutine start, not through this method — that call site is
// POST-construction (it runs on the node's own goroutine, not the single-threaded wiring
// pass) and is left as-is.
func (m *NodeGeometry) CopyClockSrc() {
	if m.clocks.clockSrc != nil {
		m.clocks.clk = m.clocks.clockSrc.Copy()
	}
}
