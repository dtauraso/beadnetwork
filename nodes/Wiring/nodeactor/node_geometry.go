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
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
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
