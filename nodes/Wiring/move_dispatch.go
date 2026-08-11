// move_dispatch.go — decentralized node-move handling.
//
// A node-move is NOT handled by a central coordinator. Instead the loader wires each
// mover's dedicated inbound channels: there is no shared many-to-one inbox — every pair of movers that talk (a node and its
// incident edge, or two nodes joined by an edge) gets its OWN directed channel, plus every
// mover gets one dedicated "external" channel for the stdin/gesture goroutine's rare direct
// entries. The stdin reader's whole job for a move is to write each entry onto the ONE
// channel that entry addresses. No recompute, no topology logic lives in the reader.
//
// Two kinds of mover own the work, each in its own goroutine (MODEL.md: each node
// and each wire is a goroutine; geometry emission is per-goroutine —
// memory/feedback_per_goroutine_bridge.md):
//
//   - nodeMover: owns ONE node's geometry. On a move for itself it updates its own
//     held position and re-emits its own node-geometry (emitNodeGeometry).
//   - edgeMover: owns ONE edge. It holds BOTH endpoint nodeGeoms (set at load). On
//     a move of either endpoint it updates the matching endpoint, recomputes its
//     OWN segment + arc (segmentBetweenPorts/arcLengthBetweenPorts), writes them
//     onto the source Out (next placement reads them), revises any in-flight bead's
//     remaining travel on the dest wire (ReviseInFlightGeometry, fraction-preserving),
//     and emits its OWN edge geometry (tr.Geometry).
//
// This reproduces, per-goroutine, exactly what the old central applyNodeMove did in
// one stdin goroutine: same node-geometry emit, same per-edge segment/arc recompute,
// same in-flight revision, same edge-geometry emit.
//
// This file holds the MoveDispatch composer struct itself and its nodeInboxes
// sub-owner; construction lives in move_dispatch_construct.go (newMoveDispatch) and the
// public delegator API lives in move_dispatch_api.go.

package Wiring

import (
	"context"

	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	rowtables "github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	sceneswitch "github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"

	T "github.com/dtauraso/wirefold/Trace"
)

// MoveDispatch is the pure registry built at load that owns every mover and wires their
// dedicated channels together — there
// is no shared dispatch map anymore; md.mr.nodeGeoms/md.mr.edgeMovers themselves are the
// directories a mover's resolveDest closure and the external-entry helpers below look up.
type MoveDispatch struct {
	// mr owns the nodeMover/edgeMover directories (mover_registry.go). bind/centerOfNode/
	// enqueueFor/finalizeActors are called directly as md.mr.X by in-package callers; only
	// Start stays a MoveDispatch method (it also sets md.ctx).
	mr moverRegistry
	// GS owns every node/edge's load-time seed geometry (nodes/Wiring/geomseeds). Exported:
	// external callers reach it directly (md.GS.NodeSeedsFn()) instead of through
	// MoveDispatch delegator methods.
	GS geomseeds.GeomSeeds
	// tr is the trace sink (retained for trace emission; diagnostic breadcrumbs removed).
	tr *T.Trace
	// persist groups the six debounced disk persisters (move_persist.go), each nil until
	// armed by EnableViewpointPersist / EnableEditPersist after the startup seed. Grouped
	// the same way ui.vp/ui.ov/ui.gest are, so a bare test-constructed MoveDispatch only
	// has to reason about one zero-value sub-struct instead of six loose nilable fields.
	persist persisters
	// sw owns the fd-wiring for the per-node interior stream (stream_wiring.go); the
	// per-edge/per-node streams only — the dedicated VIEW stream's own fd/frame-builder/
	// tick now live on UI (nodes/Wiring/viewstate), lifted out per
	// docs/planning/gesture-actor.md. MoveDispatch's public SetEdgeStreams/SetNodeStreams
	// methods stay as thin delegators so the external API is unchanged.
	sw streamWiring
	// UI owns the camera/overlay/gesture/selection/abc-drag UI state AND the VIEW stream's
	// own write side (nodes/Wiring/viewstate — lifted out of this package per
	// docs/planning/gesture-actor.md; the earlier "uiState declined" probes,
	// docs/planning/movedispatch-decomposition.md sections 5/6b, were blocked specifically
	// by the VIEW emitter's direct field access, which moved here too). Exported: the
	// type's own surface, matching GS/RT/Scenes' existing pattern — MoveDispatch's
	// setSelectionUI/setHoverUI/sendEdgeSelect-shaped helpers still route through it via
	// bound func values (move_dispatch_api.go), since moverRegistry/layoutQuantizer/
	// edgeMover are unexported Wiring types viewstate cannot name.
	UI viewstate.UIState
	// lq owns the quantized scene-polar move math (quantized_move.go): quantizedLayout
	// gates the quantized absolute-scene-polar snap — every node is a root,
	// measured/derived about the scene center only, with no per-neighbour stored
	// coordinate (MODEL.md "the polar model"). MoveDispatch's public RootMove stays a
	// thin delegator; the several package-private quantized_move.go methods also stay
	// thin delegators so their existing in-package call sites (tests, move_dispatch_construct.go,
	// gesture.go) are unchanged.
	lq layoutQuantizer
	// Scenes owns tab switching: the anchor to persist the selection against, and the
	// quit func whose call the extension host's looping respawn follows (scene_switch.go).
	// Zero until the run's setup code arms it (Scenes.AnchorPath/Scenes.Quit, set directly
	// by runtopology/topology_run.go), so a bare test-constructed MoveDispatch can never
	// end a process. Exported: external callers reach it directly.
	Scenes sceneswitch.SceneSwitch
	// tapToInstall is a TEST-ONLY observability seam: when SetMsgTap is called (before
	// Start), this is stashed here so any nodeMover constructed AFTER that call (there
	// are none in practice — all nodeMovers are built once in newMoveDispatch, before
	// SetMsgTap could ever run — but this keeps the two code paths symmetric and cheap)
	// also gets the tap installed at construction. The live seam every mover goroutine
	// actually fires is its OWN nm.tap field (node_mover.go), set once per mover by
	// SetMsgTap below — there is no shared/concurrently-read tap anymore: each mover
	// owns and reads only its own copy, on its own goroutine. nil in production —
	// production code never calls SetMsgTap.
	tapToInstall func(destID string, msg movemsg.Msg)
	// ctx is the process-lifetime context, captured in Start. sendMove (the bare
	// blocking directory send, used by external entry points like RootMove with
	// no owning mover goroutine to thread a ctx from) selects on ctx.Done() so a
	// send into a full inbox aborts on shutdown instead of leaking the calling
	// goroutine forever. nil in tests that build a bare MoveDispatch without
	// calling Start — sendMove treats a nil ctx as "no cancellation available"
	// and falls back to the plain blocking send (matches prior test behavior).
	ctx context.Context

	// RT owns the three row-identity tables (nodes/Wiring/rowtables). Exported: external
	// callers (root package) reach it directly (md.RT.NodeRowFor(...)) instead of through
	// MoveDispatch delegator methods.
	RT rowtables.RowTables

	// inboxes owns the directories of per-node dedicated channels a kind can claim for
	// itself at build time — see nodeInboxes' own doc comment.
	inboxes nodeInboxes
}

// nodeInboxes holds the DIRECTORIES OF DEDICATED PER-NODE CHANNELS that kinds claim for
// themselves at build time, one map per thing a node can be sent.
//
// It exists as an owner type rather than as loose MoveDispatch fields because that is what
// the composer rule asks for (check-composer-fields.sh): a new thing a node can be
// sent is a new entry HERE, and the composer's field count does not move. The two maps
// share one lifecycle exactly — written once per entry on the single-threaded build path
// (buildNodes, via BuildArgs), before any goroutine runs, and never touched again — so
// after build they are read-only lookup tables, which is what lets the stdin-reader
// goroutine read them without coordination.
type nodeInboxes struct {
	// tiltEdit holds, for each node id whose OWN kind claimed BuildArgs.TiltEditIn (PairNode —
	// the only kind that owns its tilt index independently), that node's dedicated inbound
	// channel for a panel-driven tilt-angle click. A node id with no entry is a kind that
	// still routes tiltVectorAngle straight to its mover (applyUpdateTiltVector's fallback,
	// stdin_reader.go). Read only by sendTiltEdit.
	tiltEdit map[string]chan movemsg.TiltEditMsg

	// lattice holds, for each node id whose own kind claimed BuildArgs.LatticeIn (PairNode —
	// the only kind that owns a lattice), that node's dedicated inbound channel for a
	// scene-level point-count change. Read only by BroadcastLatticePoints, which sends to
	// every entry: the count is one scene-wide setting, so unlike a tilt edit it is
	// addressed to no single node.
	lattice map[string]chan int32
}

// broadcastLatticePoints sends a new lattice point count to every registered node's own
// dedicated LatticeIn channel, non-blocking latest-wins (drain-then-send, the same shape
// as wire.SendSpeedNonBlocking/SendLatestNonBlocking, just over a chan int32 instead of
// chan float64/int64) so a node that is mid-cycle never blocks the sender.
func (ib *nodeInboxes) broadcastLatticePoints(points int32) {
	for _, ch := range ib.lattice {
		scenepersist.SendLatticePointsNonBlocking(ch, points)
	}
}
