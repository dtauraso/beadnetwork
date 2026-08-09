// node_geometry.go — nodeGeometry: what a node's geometry actually IS (position, reach,
// kind, neighbour centres, outgoing wires, its own retry queue, tilt/received mirror
// fields, persist root, trace, dedicated stream writer) and every method that acts on it
// (task/pair-node-owns-itself — separating STATE+BEHAVIOUR from the ACTOR that drives it).
//
// A nodeGeometry is NOT a goroutine and NOT an inbox-draining actor by itself — it owns its
// own dedicated inbound channels (extIn/neighborIn) because a node's messages are part of
// what a node IS, but nothing here ever blocks on them or loops. Two different things drive
// a nodeGeometry's per-cycle work today:
//
//   - a RING node's dedicated nodeMover actor (node_mover.go: its own goroutine, launched
//     by moverRegistry.start), which owns a *nodeGeometry and paces it on its own clock; or
//   - a PAIR node's own kind goroutine (PairNode, via BuildArgs.ClaimSelfDrive), which
//     owns a *nodeGeometry DIRECTLY — there is no nodeMover for it at all, no second
//     goroutine, nothing to skip launching.
//
// Either way exactly ONE goroutine ever touches a given nodeGeometry's mutable state — the
// invariant node_mover.go's own doc comments state throughout, unchanged by this split.
package Wiring

import (
	"encoding/binary"
	"fmt"
	"math"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// nodeGeometry owns one node's geometry, its own inbound channel set (extIn + one per
// neighbor — there is no single shared inbox), and its own outbound retry queue. On a move
// for itself it updates its held position and re-emits its node-geometry.
//
// It is a thin COMPOSER (same pattern as MoveDispatch, node_move.go): each concern is a
// NAMED sub-object declared in node_geometry_parts.go and accessed explicitly
// (m.ui.selected), never embedded — embedding would keep the old flat 46-field namespace
// and hide which owner a field belongs to. New state belongs on (or as) one of those
// owners, not as another loose field here. Guard: tools/check-composer-fields.sh.
type nodeGeometry struct {
	id   string
	geom nodeGeom
	// persistRoot is the tree root this node writes its OWN per-node files (position.json
	// — quant_offset_persist.go; port anchor files — scene_anchor_persist.go) into. Set
	// once, for every node, by MoveDispatch.EnableEditPersist after the startup seed.
	// Empty ("") means unarmed — bare test construction, or a MoveDispatch built without
	// EnableEditPersist — and every persist* method below is a no-op. This node's own
	// goroutine (whichever one drives it) reads it only from its own persist* methods, so
	// no synchronization is needed even though every node shares the same
	// EnableEditPersist call that sets it (a plain string write before any driving
	// goroutine starts).
	persistRoot string
	selfKind    string
	// quantOffset is THIS node's own quantized polar offset (iTheta,iPhi,iR + step
	// constants) about the scene center. Seeded at load (buildMoveDispatch) from the
	// computed/persisted offset, then mutated ONLY by this node's own commit path
	// (commitNodeMoveCommon, called from this node's own driving goroutine via
	// commitLocal) — single-writer, no map, no race.
	quantOffset quantizedOffset
	tr          *T.Trace

	// There is no geomMu. m.geom (port_geometry.go) splits into an embedded, write-once
	// nodeIdentity (Kind/Label/R/SceneCenter — set once at construction in loader.go,
	// grepped clean of any later write anywhere in this package) and MUTABLE state
	// (ScenePolar/HasPos/ReachR) written only by applyCenter. Every writer AND every
	// reader of the mutable part — applyCenter, emitGeometry's full-struct copy — runs
	// exclusively on this node's OWN driving goroutine (whichever one that is), so there
	// is never more than one goroutine touching that memory. The one cross-goroutine
	// reader, MoveDispatch.NodeKind (node_move.go), called from the gesture/stdin-reader
	// goroutine, reads ONLY nm.geom.Kind — a field on the embedded nodeIdentity, which no
	// writer here ever touches.
	//
	// CHECKED BY CODE: TestNodeKindConcurrentWithApplyCenterUnderRace
	// (node_mover_geom_race_test.go) drives NodeKind's reader loop and applyCenter's
	// writer loop concurrently under -race, as a standing regression check that the split
	// holds.

	// msg owns this node's dedicated inbound channels, its outbound retry queue and the
	// routing closures it hands a moveMsg to (nodeMessaging, node_geometry_parts.go).
	msg nodeMessaging
	// clocks owns the clock source this node copies from once and its own copy
	// (nodeClocks).
	clocks nodeClocks
	// stream owns this node's dedicated per-node content-buffer stream: its fd, its row,
	// its kind column and its frame packer (nodeStream).
	stream nodeStream
	// ui owns this node's OWN selection/hover bytes — per-owner, no shared or republished
	// map (nodeUI).
	ui nodeUI
	// tilt owns this node's tilt/received-vector mirror indices and its lattice size
	// (nodeTilt).
	tilt nodeTilt
	// readout owns the two pair vector-exchange span counters (pairReadout).
	readout pairReadout
	// outs owns this node's outgoing targets, paced wires, Outs and step channels
	// (nodeOuts).
	outs nodeOuts
	// topo owns this node's own view of its adjacency: incident edges, partner centres and
	// kinds, mutual targets, neighbour row lookup (neighborTopology).
	topo neighborTopology
	// flags owns the two scene-wide ring-axis drawing choices this node applies to its own
	// frame (sceneFlags).
	flags sceneFlags
	// beads owns this node's placeholder chain-bead actors and their tick source
	// (nodeBeads).
	beads nodeBeads
}

// newNodeGeometry constructs one node's geometry — no actor, no goroutine. Whoever drives
// it (a ring's nodeMover, or a pair kind's own goroutine via ClaimSelfDrive) copies
// clockSrc into clk once, at its own start.
func newNodeGeometry(id string, geom nodeGeom, tr *T.Trace, clockSrc wire.Clock) *nodeGeometry {
	ng := &nodeGeometry{
		id: id, geom: geom, tr: tr,
		msg: nodeMessaging{
			extIn:      make(chan moveMsg, moverInboxDepth),
			neighborIn: map[string]chan moveMsg{},
			centerOut:  make(chan vec3, 1),
		},
		topo:   neighborTopology{partnerCenters: map[string]vec3{}},
		clocks: nodeClocks{clockSrc: clockSrc, clk: wire.NewRealClock()},
		tilt:   nodeTilt{latticePoints: FullTurnThetaIdx},
	}
	// Self-seed centerOut with the initial geometry (even when !HasPos, in which case
	// nodeWorldPos falls back to the origin) so the dispatch goroutine's first drain
	// always finds a valid center.
	ng.msg.centerOut <- nodeWorldPos(geom)
	// Production-only hook: arms the bead-actor path in chainBeads/reconcileBeadChain
	// (bead_chain.go). Bare `&nodeGeometry{...}` test literals never call
	// newNodeGeometry, so beadTickFn stays nil there and chainBeads' pure-function tests
	// never touch a live TickBroadcaster goroutine.
	ng.beads.beadTickFn = wire.NewTickChan
	return ng
}

// handle applies one move to this node: update held position, re-emit node-geometry.
func (m *nodeGeometry) handle(msg moveMsg) {
	if msg.NodeID != m.id {
		return
	}
	if msg.Kind == moveMsgKindCenter {
		// This node is the SOLE writer of its own position (single-writer by
		// construction — this is the only path that mutates it). A Center payload is
		// the flat absolute-scene-polar drag write from fanCenters: apply it via
		// applyCenter, which also re-emits. A nil Center is fanCenters' PARTNER
		// re-emit (a neighbor whose OWN center is unchanged, only asked to re-emit so
		// any reader of its geometry sees a consistent frame) — no mutation, just
		// re-emit.
		if msg.Center != nil {
			m.applyCenter(*msg.Center, msg.ReachR)
			return
		}
		if m.tr != nil {
			m.emitGeometry()
		}
		return
	}
	if msg.Kind == moveMsgKindDrag {
		// Owner-goroutine drag entry (generalized to EVERY node so no node's quantized
		// offset is ever touched by a foreign goroutine): commit this node's OWN new
		// position via the local (synchronous-snap-publish) commit path. A drag is
		// always a FREE move now -- there is no equal-radii solve and no propagation
		// past this node's own commit.
		newPos := msg.Target
		if m.msg.commitLocal != nil {
			m.msg.commitLocal(m.id, newPos)
		}
		if m.tr != nil {
			m.tr.Breadcrumb("drag.commit", m.id, "", fmt.Sprintf("newPos=(%.4f,%.4f,%.4f)", newPos.X, newPos.Y, newPos.Z))
			// Structured buffer counterpart, riding this node's own dedicated
			// stream frame (emitGeometry's own next emit already fires from
			// commitLocal above, so this rides as a distinct events-only-shaped
			// write here rather than waiting on that one).
			m.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbDragCommit, Debug: 1,
				NodeRow: m.stream.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				X: newPos.X, Y: newPos.Y, Z: newPos.Z,
			}})
		}
		return
	}
	if msg.Kind == moveMsgKindDragStart {
		m.startBeadDrag()
		return
	}
	if msg.Kind == moveMsgKindDragEnd {
		// "done dragging" is not optional (PLAN.md): sent from gesture.go's
		// gestPointerUp on EVERY path a drag can end by (including one abandoned
		// without a clean pointer-move first — see moveMsgKindDragEnd's own doc
		// comment), so no chain bead this node woke is ever left on machine time.
		m.endBeadDrag()
		return
	}
	if msg.Kind == moveMsgKindSelect {
		if msg.Bool {
			m.ui.selected = 1
		} else {
			m.ui.selected = 0
		}
		return
	}
	if msg.Kind == moveMsgKindHover {
		if msg.Bool {
			m.ui.hovered = 1
			m.ui.hoverPort = msg.Port
			m.ui.hoverIsInput = msg.IsInput
		} else {
			m.ui.hovered = 0
			m.ui.hoverPort = ""
			m.ui.hoverIsInput = false
		}
		return
	}
	if msg.Kind == moveMsgKindLatched {
		if msg.Bool {
			m.ui.latchedSel = 1
		} else {
			m.ui.latchedSel = 0
		}
		return
	}
	if msg.Kind == moveMsgKindTiltVectorAngle {
		// Adjust THIS node's own vector-direction index by one TiltVectorAngleStep click —
		// index arithmetic only (memory/feedback_abc_times_constant_not_rederive.md), no
		// trig here. Persisted immediately to this node's OWN file (persistTiltVectorAngle,
		// quant_offset_persist.go) and re-emitted so the panel's read-only reflect and
		// the drawn arrow both pick up the change on the next frame.
		delta := int32(-1)
		if msg.Bool {
			delta = 1
		}
		m.tilt.topTiltVectorThetaIdx += delta
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: this path only runs for a kind that has NOT claimed BuildArgs.TiltEditIn
		// (every kind except PairNode today — see moveMsgKindTiltVectorAngle's own doc
		// comment and applyUpdateTiltVector's fallback, stdin_reader.go). PairNode's own
		// tilt-panel edits are routed to its OWN goroutine instead (TiltEditIn), which
		// applies the click, syncs this value back via PairNodeSelf.SetTiltIndex, AND places
		// "the kick" bead on its own Out directly — none of that happens here anymore.
		return
	}
	if msg.Kind == moveMsgKindTiltVectorReset {
		// Return THIS node's own vector direction to the start position — both indices to
		// 0, the documented default (tilt vector at world +y). No bead: this is a
		// stop-and-return, not a kick. Persisted immediately, same as an adjust.
		m.tilt.topTiltVectorThetaIdx = 0
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: same split as moveMsgKindTiltVectorAngle — this path only runs for a kind
		// that has NOT claimed BuildArgs.TiltEditIn. PairNode routes a reset through
		// its own TiltEditIn/TiltEditMsg.Reset instead.
		return
	}
	// moveMsgKindTiltIndexSync/ReceivedVectorSync/BeadClear are GONE
	// (task/pair-node-owns-itself): a pair node (PairNode) owns this geometry
	// directly (PairNodeSelf, pair_node_self.go), so what used to be a one-way
	// notification message to itself is now a plain method call on the same
	// goroutine — see PairNodeSelf.SetTiltIndex/SetReceivedVector/ClearOutBeads,
	// which apply exactly what this handle() branch used to apply.
	if msg.Kind == moveMsgKindNeighborCenter {
		// Delivery-mechanism push (see applyCenter/partnerCenters' doc comments): a
		// direct neighbor's OWN center just changed. Store it in THIS node's owned
		// partnerCenters map (write, own goroutine only) and re-emit THIS node's own
		// geometry so its aimed ports pick up the fresh partner center — same value,
		// same effect as the old cross-goroutine snap read, just message-delivered.
		// ONE HOP ONLY: this node's own center did NOT change, so it must never push
		// a NeighborCenter of its own onward from here (no cascade past this point).
		if m.topo.partnerCenters == nil {
			m.topo.partnerCenters = map[string]vec3{}
		}
		m.topo.partnerCenters[msg.SenderID] = msg.FromCenter
		if m.tr != nil {
			// DIAGNOSTIC ONLY (task/log-node4-chain-aim): records that this node's own
			// goroutine received a neighbor-center push, and from whom, so a drag-time
			// trace can show whether/when it arrives relative to this node's own emits.
			value := fmt.Sprintf("sender=%s center=(%.4f,%.4f,%.4f)", msg.SenderID, msg.FromCenter.X, msg.FromCenter.Y, msg.FromCenter.Z)
			m.tr.Breadcrumb("neighbor-center-recv", m.id, msg.SenderID, value)
			senderRow := int32(-1)
			if m.topo.nodeRowFor != nil {
				if r, ok := m.topo.nodeRowFor(msg.SenderID); ok {
					senderRow = r
				}
			}
			m.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbNeighborCenterRecv, Debug: 1,
				NodeRow: m.stream.nodeRow, PortRow: -1, TargetRow: senderRow, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Text: value,
			}})
			m.emitGeometry()
		}
		return
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

// PerpendicularThetaIdx is the topTiltVectorThetaIdx value at which the tilt vector is exactly
// perpendicular to world +y: CurveParamTiltVectorAngleStep is π/12 (15°), and π/2 (90°) is
// exactly 6 steps. Comparing to this INTEGER is what makes the straightening loop's stop
// condition exact — cos(π/2) in float64 is ~6.1e-17, so a literal float dot==0 test would
// never fire (memory/feedback_abc_times_constant_not_rederive.md: index arithmetic, trig
// only at the cartesian/polar boundary). Exported (capitalized) so PairNode's own
// goroutine — which now runs the straightening rule itself, per-package — can compare
// against it without duplicating the constant; the rule itself no longer lives here (see
// nodes/PairNode/node.go).
//
// dot(tilt, coplanarNormal) == 0 is decided as thetaIdx == PerpendicularThetaIdx, not by
// computing an actual float dot product. STATE THE ASSUMPTION THAT MAKES THE SHORTCUT
// VALID: the tilt vector's in-plane angle IS its θ index only because, for this scene, the
// ring plane holds world +y and θ is measured from +y, so the two coincide (see
// topTiltVectorThetaIdx's own doc comment and the CoplanarNormal/UpAxis derivations in
// writeStreamFrame above). A scene whose ring plane does NOT contain +y breaks that
// coincidence — θ would then measure something unrelated to the coplanar normal, and the
// rule would need to compare an actual dot(tilt, coplanarNormal) via the two integer
// indices' angles converted through anglesToWorldOffset, not thetaIdx alone.
const PerpendicularThetaIdx int32 = 6

// applyCenter is the SOLE WRITE of this node's center/reach. It is called ONLY from this
// node's own driving goroutine (handle's moveMsgKindCenter case, driven by fanCenters
// below), which is what makes that one goroutine the exclusive writer of m.geom. It sets
// the held polar position, pushes the fresh center to the dispatch goroutine's owned
// center mirror (m.msg.centerOut, latest-wins — see its doc comment) and to every direct
// neighbor's partnerCenters map (below), and re-emits this node's live geometry.
func (m *nodeGeometry) applyCenter(center vec3, reach float64) {
	setNodeWorld(&m.geom, center)
	m.geom.ReachR = reach
	// Latest-wins non-blocking push onto centerOut: drain any stale unread value first
	// so the slot always ends up holding the newest center, never blocking this
	// goroutine even if the dispatch goroutine hasn't drained the previous push yet.
	select {
	case <-m.msg.centerOut:
	default:
	}
	select {
	case m.msg.centerOut <- center:
	default:
	}
	// Push this fresh center to every direct neighbor (nm.msg.neighborIn's key set — one
	// hop, no cascade) so each neighbor's OWN partnerCenters map picks it up via
	// moveMsgKindNeighborCenter (handle, below). Routed through m.msg.sendMove (this
	// node's own retry queue), same as every other fan-out this file makes, so a
	// momentarily-full neighbor inbox is retried, never dropped or blocking. Sent
	// BEFORE this same commit's broadcastToEdgesAndPartners nil-Center re-emit (called
	// right after applyCenter by every live caller), so per-destination FIFO delivers
	// this push first and the re-emit always sees the just-pushed center.
	for neighborID := range m.msg.neighborIn {
		m.msg.sendMove(neighborID, moveMsg{Kind: moveMsgKindNeighborCenter, NodeID: neighborID,
			SenderID: m.id, FromCenter: center})
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

// emitGeometry re-emits this node's authoritative geometry (center, radius, ring
// normals — no port geometry: a port carries none, docs/channels-not-ports.md).
// This method and applyCenter both run on this node's own driving goroutine only, so a
// plain field read here can never race a concurrent writer.
func (m *nodeGeometry) emitGeometry() {
	// Dedicated per-node stream (see streamOut's doc comment): write this node's own
	// combined frame immediately on a geometry change, in addition to the tick-driven
	// write in the driving loop's own per-cycle write. NodeGeometry rides THIS frame's
	// own EVENTS section (fully decentralized — it never rides the VIEW stream's
	// fallback bucket) — this node is the sole owner of its own geometry, so it resolves
	// its own NodeRow at the call site (owner_events.go) rather than routing through a
	// shared accumulator.
	m.writeStreamFrame([]wire.RowEvent{{
		Kind: T.KindNodeGeometry, NodeRow: m.stream.nodeRow,
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}})
}

// writeStreamFrame packs and writes this node's combined per-fd frame (center/radius/
// ring-normals + ports + label + selection-UI columns) to its OWN dedicated fd
// (streamOut). No-op when streamOut is nil (the fallback — see its doc comment) or
// buildFrame was never injected (bare test construction). Called only by this node's own
// driving goroutine, reading m.geom. events carries whatever this call's caller wants
// riding this frame's trailing EVENTS section (nil from a plain tick-driven write).
func (m *nodeGeometry) writeStreamFrame(events []wire.RowEvent) {
	if !m.stream.streamOut.Ok() || m.stream.buildFrame == nil {
		return
	}
	// INVARIANT: a node carries only its OWN events on its OWN dedicated stream. This is
	// the per-goroutine bridge stated in CLAUDE.md's "Bridge surface" and in
	// memory/feedback_no_single_writer_bridge.md + memory/feedback_per_goroutine_bridge.md,
	// and until now it was enforced by prose alone. NodeRow is the ownership column; a
	// FOREIGN node is referenced through TargetRow (see quantized_move.go's abc-drag
	// breadcrumb, which sets NodeRow: nm.stream.nodeRow and TargetRow: the other node). Violating
	// it produces a frame the TS side decodes onto the wrong row — a silently wrong scene
	// that still renders, which is the expensive failure this panic converts into a cheap
	// one. Placed AFTER the nil guard on purpose: bare geometries built in tests never
	// reach the pack path, and nodeRow is seeded alongside streamOut (stream_wiring.go),
	// so any frame that gets here has a real row.
	for _, e := range events {
		if e.NodeRow != m.stream.nodeRow {
			panic(fmt.Sprintf(
				"nodeGeometry.writeStreamFrame: node %q (row %d) is carrying a %s event for row %d on its OWN dedicated stream — NodeRow is an ownership claim, not a reference; a foreign node belongs in TargetRow",
				m.id, m.stream.nodeRow, e.Kind, e.NodeRow))
		}
	}
	center := nodeWorldPos(m.geom)
	sphereR := effectiveRadius(m.geom)
	// This node's own local-frame pole: its own scene-polar direction reversed, so the frame
	// points back at the scene centre (Buffer/layout.go PoleTheta/PolePhi). Derived here from
	// m.geom.ScenePolar — this node's own coordinate, on this node's own goroutine, no
	// neighbour read. Before HasPos there is no direction yet, so the frame stays world +y.
	var poleTheta, polePhi float64
	if m.geom.HasPos {
		poleTheta, polePhi = inwardPole(m.geom.ScenePolar)
	}
	// The DRAWN ring's axis, separate from the navigation pole above (Buffer/layout.go's
	// RingAxisTheta/RingAxisPhi). Default is the torus's own +Z normal, which draws exactly
	// as an unrotated ring did — so a scene that has not asked for anything looks unchanged.
	ringAxisTheta, ringAxisPhi := torusDefaultAxisAngles()
	// topTiltVectorLen is this node's own drawn vector, along the SAME axis as its ring, and 0
	// where a scene draws none (Buffer/layout.go's TopTiltVectorLen). It runs from the node's
	// centre to its own top, so its length IS the node's radius.
	var topTiltVectorLen float64
	if m.flags.upAxis && m.geom.HasPos && len(m.topo.partnerCenters) == 1 {
		// UPRIGHT: the ring STANDS UP along its edge — its plane holds both the edge and
		// world +y, so the node's own up-vector lies IN the ring's plane rather than
		// sticking out of a flat disc. An axis of +y itself would lie the ring flat and
		// put the vector perpendicular to it, which is the opposite arrangement.
		for _, partner := range m.topo.partnerCenters {
			if t, p, ok := uprightRingAxis(nodeWorldPos(m.geom), partner); ok {
				ringAxisTheta, ringAxisPhi = t, p
			}
		}
		topTiltVectorLen = nodeRadius(m.geom.Kind)
	} else if m.flags.coplanarEdges && m.geom.HasPos && len(m.topo.partnerCenters) == 1 {
		// COPLANAR EDGES: swing the axis off the inward pole by the smallest amount that
		// puts the edge INSIDE the ring plane — the inward pole with its along-the-edge
		// component removed. The chain, this node's torus and the beads' own tori then
		// share one plane instead of the chain running through the holes. Only for a node
		// with exactly ONE neighbour: two non-collinear edges have no common plane.
		for _, partner := range m.topo.partnerCenters {
			if t, p, ok := poleContainingEdge(poleTheta, polePhi, nodeWorldPos(m.geom), partner); ok {
				ringAxisTheta, ringAxisPhi = t, p
			}
		}
	}
	// latticeThetaStep is THIS node's own angle-per-index — 2π / latticePoints, not the
	// fixed CurveParamTiltVectorAngleStep (which stays π/12, the compile-time 24-point
	// default every OTHER conversion in this codebase still uses). A pair node's own
	// lattice size is a scene setting (Node.adoptLattice, nodes/PairNode/node.go), reported
	// here one-way via PairNodeSelf.SetLatticePoints, so the same index draws a different
	// angle depending on how many points that node's own ring currently has. Derived once
	// per frame; every conversion below reads this local rather than recomputing it.
	points := m.tilt.latticePoints
	if points == 0 {
		points = FullTurnThetaIdx
	}
	latticeThetaStep := 2 * math.Pi / float64(points)
	// topTiltVectorTheta is this node's OWN vector direction — separate from the ring
	// axis above, so a scene/user can aim a node's vector somewhere other than its ring.
	// Never a free float: index × latticeThetaStep (this node's own lattice step, above),
	// the streamed value is pure arithmetic on the integer state this node's own mover
	// holds and persists (m.tilt.topTiltVectorThetaIdx). There is no φ: every tilt vector in
	// this model is θ-only (task/drop-tilt-vector-phi).
	topTiltVectorTheta := float64(m.tilt.topTiltVectorThetaIdx) * latticeThetaStep
	// The BOTTOM TILT VECTOR: streamed straight from this node's own bottomThetaIdx,
	// decided by THIS node's OWN goroutine (a half turn in θ from its own top
	// tilt index, same rule run unmodified by both nodes of a pair — PairNode's bottomTilt)
	// and reported one-way
	// via PairNodeSelf.SetTiltIndex alongside the top and the normal. Pure mirror here, same
	// as every other index on this frame: this mover derives none of them.
	bottomTiltVectorTheta := float64(m.tilt.bottomThetaIdx) * latticeThetaStep
	// The COPLANAR NORMAL: streamed straight from this node's own normalThetaIdx,
	// which THIS node's OWN goroutine decided (a fixed +90° in θ from its
	// own tilt index, same rule run unmodified by both nodes of a pair — PairNode's
	// coplanarNormal) and reported one-way via PairNodeSelf.SetTiltIndex. This mover is a pure mirror here, same shape
	// as topTiltVectorTheta above — it derives nothing from the edge/partner.
	// Turning the tilt therefore visibly turns the drawn normal WITH it, staying 90° away,
	// instead of the normal staying fixed toward the partner while the tilt moves under it.
	coplanarNormalTheta := float64(m.tilt.normalThetaIdx) * latticeThetaStep
	// The THIRD vector: the direction last received on this node's tilt-vector channel
	// (receivedVectorThetaIdx, mirrored one-way from this node's own goroutine —
	// see the field's own doc comment). Same length-says-whether-and-how-far convention
	// as topTiltVectorLen: zero when nothing has been received yet (or a reset cleared it),
	// non-zero (this node's own radius, same as topTiltVectorLen) otherwise — so a node with
	// nothing received is distinguishable from one whose received direction happens to be
	// 0, which still streams a non-zero length.
	var receivedVectorLen float64
	var receivedVectorTheta float64
	if m.tilt.receivedVectorSet {
		receivedVectorLen = nodeRadius(m.geom.Kind)
		receivedVectorTheta = float64(m.tilt.receivedVectorThetaIdx) * latticeThetaStep
	}
	label := m.geom.Label
	if label == "" {
		label = m.id
	}
	selected, hovered, latchedSel, kindID := m.ui.selected, m.ui.hovered, m.ui.latchedSel, m.stream.kindID
	// This node's own placeholder chain beads, node-local (chain_beads.go). Computed here
	// on this node's own goroutine from its own center + its own partnerCenters map — no
	// cross-goroutine position read.
	chainOX, chainOY, chainOZ, chainLit, chainLitVal, chainBreadcrumbs := m.chainBeads()
	if len(chainBreadcrumbs) > 0 {
		// DIAGNOSTIC ONLY (task/log-node4-chain-aim): chainBeads' own "chain-aim" events,
		// appended here rather than sent via a nested writeStreamFrame call from inside
		// chainBeads (which would recurse back into chainBeads — see that function's doc
		// comment on its breadcrumbs return value).
		events = append(events, chainBreadcrumbs...)
	}
	// nodeID is this node's own numeric identity: ROW ID = NODE ID - 1 (enforced at load,
	// persistence-ownership.md), so it is m.stream.nodeRow+1 by construction — not re-derived by any
	// offline rule the decoder also has to apply, it travels with the frame.
	frame := m.stream.buildFrame(NodeFrameInput{
		Tick:                  uint32(m.clocks.clk.Tick()),
		NodeRow:               m.stream.nodeRow,
		NodeID:                m.stream.nodeRow + 1,
		CX:                    float32(center.X),
		CY:                    float32(center.Y),
		CZ:                    float32(center.Z),
		Radius:                float32(nodeRadius(m.geom.Kind)),
		SphereR:               float32(sphereR),
		VRX:                   verticalRingNormalX,
		VRY:                   verticalRingNormalY,
		VRZ:                   verticalRingNormalZ,
		FRX:                   flatRingNormalX,
		FRY:                   flatRingNormalY,
		FRZ:                   flatRingNormalZ,
		PoleTheta:             float32(poleTheta),
		PolePhi:               float32(polePhi),
		RingAxisTheta:         float32(ringAxisTheta),
		RingAxisPhi:           float32(ringAxisPhi),
		TopTiltVectorLen:      float32(topTiltVectorLen),
		TopTiltVectorTheta:    float32(topTiltVectorTheta),
		BottomTiltVectorTheta: float32(bottomTiltVectorTheta),
		CoplanarNormalTheta:   float32(coplanarNormalTheta),
		ReceivedVectorLen:     float32(receivedVectorLen),
		ReceivedVectorTheta:   float32(receivedVectorTheta),
		Selected:              selected,
		KindID:                kindID,
		Hovered:               hovered,
		LatchedSel:            latchedSel,
		LatticePoints:         uint8(points),
		RoundsToParallel:      m.readout.roundsToParallel,
		MsgsToParallel:        m.readout.msgsToParallel,
		Label:                 label,
		ChainBeadOX:           chainOX,
		ChainBeadOY:           chainOY,
		ChainBeadOZ:           chainOZ,
		ChainBeadLit:          chainLit,
		ChainBeadLitValue:     chainLitVal,
		Events:                events,
	})
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	// Fire-and-forget, same reasoning throughout this bridge: no delivery
	// guarantee on this channel, errors ignored.
	_, _ = m.stream.streamOut.Write(hdr[:])
	_, _ = m.stream.streamOut.Write(frame)
}

// flushPending retries every message in m.msg.pending in order, attempting a non-blocking
// send to its destination's inbox. A destination whose channel is momentarily full
// stays in the queue (retried next call) — and so does every LATER item addressed to
// that SAME destination, even if its own channel isn't full, so per-destination FIFO
// is preserved (a retained item is never overtaken by a newer one to the same
// destination). An item whose destination doesn't resolve (unknown id) is dropped,
// matching the old deliverMove no-op for an unknown id. Called only from m's own
// driving goroutine (sendMove, at enqueue time, and the driving loop, every cycle).
func (m *nodeGeometry) flushPending() {
	if len(m.msg.pending) == 0 || m.msg.resolveDest == nil {
		return
	}
	blocked := map[string]bool{}
	kept := m.msg.pending[:0]
	for _, item := range m.msg.pending {
		if blocked[item.destID] {
			kept = append(kept, item)
			continue
		}
		ch, ok := m.msg.resolveDest(item.destID)
		if !ok {
			continue
		}
		select {
		case ch <- item.msg:
		default:
			blocked[item.destID] = true
			kept = append(kept, item)
		}
	}
	m.msg.pending = kept
}
