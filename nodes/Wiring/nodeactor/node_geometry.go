// node_geometry.go — NodeGeometry: what a node's geometry actually IS (position, reach,
// kind, neighbour centres, outgoing wires, its own retry queue, tilt/received mirror
// fields, persist root, trace, dedicated stream writer), its constructor, and its message
// dispatch (task/pair-node-owns-itself — separating STATE+BEHAVIOUR from the ACTOR that
// drives it).
//
// A NodeGeometry is NOT a goroutine and NOT an inbox-draining actor by itself — it owns its
// own dedicated inbound channels (extIn/neighborIn) because a node's messages are part of
// what a node IS, but nothing here ever blocks on them or loops. Two different things drive
// a NodeGeometry's per-cycle work today:
//
//   - a RING node's dedicated NodeMover actor (node_mover.go: its own goroutine, launched
//     by package Wiring's moverRegistry.start), which owns a *NodeGeometry and paces it on
//     its own clock; or
//   - a PAIR node's own kind goroutine (PairNode, via BuildArgs.ClaimSelfDrive), which
//     owns a *NodeGeometry DIRECTLY — there is no NodeMover for it at all, no second
//     goroutine, nothing to skip launching.
//
// Either way exactly ONE goroutine ever touches a given NodeGeometry's mutable state — the
// invariant node_mover.go's own doc comments state throughout, unchanged by this split.
//
// This package (nodeactor) is the per-node ACTOR moved out of package Wiring
// (docs/planning/movedispatch-decomposition.md §20 — the same package-boundary move §17
// made for edgeMover). Package Wiring reaches this type ONLY through the exported surface
// declared here and in node_geometry_wire.go/node_geometry_accessors.go; every field stays
// unexported, and every channel this type owns stays unexported behind a method that
// closes over it (never a raw channel export — the same rule §17 held for edgeMover).
//
// The rest of what a NodeGeometry DOES lives one job per file, alongside this one: its sole
// center/reach write is node_geometry_center.go, its dedicated-stream frame packing is
// node_geometry_stream.go, and its outbound retry-queue drain is node_geometry_retry.go.
package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// NodeGeometry owns one node's geometry, its own inbound channel set (extIn + one per
// neighbor — there is no single shared inbox), and its own outbound retry queue. On a move
// for itself it updates its held position and re-emits its node-geometry.
//
// It is a thin COMPOSER (same pattern as package Wiring's MoveDispatch): each concern is a
// NAMED sub-object declared in node_geometry_parts.go and accessed explicitly
// (m.ui.selected), never embedded — embedding would keep a flat field namespace and hide
// which owner a field belongs to. New state belongs on (or as) one of those owners, not as
// another loose field here. Guard: tools/network/structure/check-composer-fields.sh.
type NodeGeometry struct {
	id   string
	geom nodegeom.NodeGeom
	// persistRoot is the tree root this node writes its OWN per-node files (position.json
	// — quant_offset_persist.go; port anchor files — scene_anchor_persist.go, package
	// Wiring) into. Set once, for every node, by package Wiring's MoveDispatch.
	// EnableEditPersist after the startup seed, via SetPersistRoot (node_geometry_wire.go).
	// Empty ("") means unarmed — bare test construction, or a MoveDispatch built without
	// EnableEditPersist — and every persist* method below is a no-op. This node's own
	// goroutine (whichever one drives it) reads it only from its own persist* methods, so
	// no synchronization is needed even though every node shares the same
	// EnableEditPersist call that sets it (a plain string write before any driving
	// goroutine starts).
	persistRoot string
	selfKind    string
	// quantOffset is THIS node's own quantized polar offset (iTheta,iPhi,iR + step
	// constants) about the scene center. Seeded at load (SetQuantOffset) from the
	// computed/persisted offset, then mutated ONLY by this node's own commit path
	// (CommitQuantOffset, called from this node's own driving goroutine) — single-writer,
	// no map, no race.
	quantOffset quantoffset.QuantizedOffset
	tr          *T.Trace

	// There is no geomMu. m.geom (nodegeom/port_geometry.go) splits into an embedded, write-once
	// NodeIdentity (Kind/Label/R/SceneCenter — set once at construction in package Wiring's
	// loader.go, grepped clean of any later write anywhere outside applyCenter) and MUTABLE
	// state (ScenePolar/HasPos/ReachR) written only by ApplyCenter. Every writer AND every
	// reader of the mutable part — ApplyCenter, emitGeometry's full-struct copy — runs
	// exclusively on this node's OWN driving goroutine (whichever one that is), so there
	// is never more than one goroutine touching that memory. The one cross-goroutine
	// reader, this type's own Kind() accessor (called from package Wiring's dispatch/
	// gesture goroutine), reads ONLY the embedded NodeIdentity's Kind field, which no
	// writer here ever touches.
	//
	// CHECKED BY CODE: TestNodeKindConcurrentWithApplyCenterUnderRace
	// (node_mover_geom_race_test.go) drives Kind's reader loop and ApplyCenter's
	// writer loop concurrently under -race, as a standing regression check that the split
	// holds.

	// msg owns this node's dedicated inbound channels, its outbound retry queue and the
	// routing closures it hands a movemsg.Msg to (nodeMessaging, node_geometry_parts.go).
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

// NewNodeGeometry constructs one node's geometry — no actor, no goroutine. Whoever drives
// it (a ring's NodeMover, or a pair kind's own goroutine via ClaimSelfDrive) copies
// clockSrc into clk once, at its own start.
func NewNodeGeometry(id string, geom nodegeom.NodeGeom, tr *T.Trace, clockSrc clock.Clock) *NodeGeometry {
	ng := &NodeGeometry{
		id: id, geom: geom, tr: tr,
		msg: nodeMessaging{
			extIn:      make(chan movemsg.Msg, inboxDepth),
			neighborIn: map[string]chan movemsg.Msg{},
			centerOut:  make(chan vec3, 1),
		},
		topo:   neighborTopology{partnerCenters: map[string]vec3{}},
		clocks: nodeClocks{clockSrc: clockSrc, clk: clock.NewRealClock()},
		tilt:   nodeTilt{latticePoints: tiltvector.FullTurnThetaIdx},
	}
	// Self-seed centerOut with the initial geometry (even when !HasPos, in which case
	// nodeWorldPos falls back to the origin) so the dispatch goroutine's first drain
	// always finds a valid center.
	ng.msg.centerOut <- nodegeom.NodeWorldPos(geom)
	// Production-only hook: arms the bead-actor path in chainBeads/reconcileBeadChain
	// (bead_chain.go). Bare `&NodeGeometry{...}` test literals never call
	// NewNodeGeometry, so beadTickFn stays nil there and chainBeads' pure-function tests
	// never touch a live TickBroadcaster goroutine.
	ng.beads.beadTickFn = clock.NewTickChan
	return ng
}

// handle applies one move to this node: update held position, re-emit node-geometry.
func (m *NodeGeometry) handle(msg movemsg.Msg) {
	if msg.NodeID != m.id {
		return
	}
	if msg.Kind == movemsg.KindCenter {
		// This node is the SOLE writer of its own position (single-writer by
		// construction — this is the only path that mutates it). A Center payload is
		// the flat absolute-scene-polar drag write from fanCenters: apply it via
		// ApplyCenter, which also re-emits. A nil Center is fanCenters' PARTNER
		// re-emit (a neighbor whose OWN center is unchanged, only asked to re-emit so
		// any reader of its geometry sees a consistent frame) — no mutation, just
		// re-emit.
		if msg.Center != nil {
			m.ApplyCenter(*msg.Center, msg.ReachR)
			return
		}
		if m.tr != nil {
			m.emitGeometry()
		}
		return
	}
	if msg.Kind == movemsg.KindDrag {
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
	if msg.Kind == movemsg.KindDragStart {
		m.startBeadDrag()
		return
	}
	if msg.Kind == movemsg.KindDragEnd {
		// "done dragging" is not optional (PLAN.md): sent from gesture.go's
		// gestPointerUp on EVERY path a drag can end by (including one abandoned
		// without a clean pointer-move first — see movemsg.KindDragEnd's own doc
		// comment), so no chain bead this node woke is ever left on machine time.
		m.endBeadDrag()
		return
	}
	if msg.Kind == movemsg.KindSelect {
		if msg.Bool {
			m.ui.selected = 1
		} else {
			m.ui.selected = 0
		}
		return
	}
	if msg.Kind == movemsg.KindHover {
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
	if msg.Kind == movemsg.KindLatched {
		if msg.Bool {
			m.ui.latchedSel = 1
		} else {
			m.ui.latchedSel = 0
		}
		return
	}
	if msg.Kind == movemsg.KindTiltVectorAngle {
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
		// (every kind except PairNode today — see movemsg.KindTiltVectorAngle's own doc
		// comment and applyUpdateTiltVector's fallback, package Wiring's stdin_reader.go).
		// PairNode's own tilt-panel edits are routed to its OWN goroutine instead
		// (TiltEditIn), which applies the click, syncs this value back via
		// PairNodeSelf.SetTiltIndex, AND places "the kick" bead on its own Out directly
		// — none of that happens here anymore.
		return
	}
	if msg.Kind == movemsg.KindTiltVectorReset {
		// Return THIS node's own vector direction to the start position — both indices to
		// 0, the documented default (tilt vector at world +y). No bead: this is a
		// stop-and-return, not a kick. Persisted immediately, same as an adjust.
		m.tilt.topTiltVectorThetaIdx = 0
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: same split as movemsg.KindTiltVectorAngle — this path only runs for a kind
		// that has NOT claimed BuildArgs.TiltEditIn. PairNode routes a reset through
		// its own TiltEditIn/movemsg.TiltEditMsg.Reset instead.
		return
	}
	// movemsg.KindTiltIndexSync/ReceivedVectorSync/BeadClear are GONE
	// (task/pair-node-owns-itself): a pair node (PairNode) owns this geometry
	// directly (PairNodeSelf, pair_node_self.go), so what used to be a one-way
	// notification message to itself is now a plain method call on the same
	// goroutine — see PairNodeSelf.SetTiltIndex/SetReceivedVector/ClearOutBeads,
	// which apply exactly what this handle() branch used to apply.
	if msg.Kind == movemsg.KindNeighborCenter {
		// Delivery-mechanism push (see ApplyCenter/partnerCenters' doc comments): a
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
