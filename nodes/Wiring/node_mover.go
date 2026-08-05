// node_mover.go — the per-node mover actor type split out of node_move.go. Pure
// move: no logic changes beyond the two-channels-no-inbox-no-blocking restructure:
// there is no shared many-to-one inbox anymore. Every pair of movers that talk gets its OWN dedicated directed channel
// (nodeMover.neighborIn, edgeMover.srcIn/dstIn — edgeMover lives in edge_mover.go), plus one dedicated "external" channel per
// mover (extIn) for the stdin/gesture goroutine's rare direct entries (drag/dragStart/
// anchor). node_move.go retains the dispatch registry (MoveDispatch) that WIRES these
// channels together at load time; this file owns the node actor (owns
// one node's geometry, its own inbound channel set, and its own outbound retry queue).
// Each mover touches MoveDispatch only via an injected enqueue func — no back-reference
// to the dispatch registry, and no shared queue/lock between movers.

package Wiring

import (
	"context"
	"encoding/binary"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"path/filepath"

	T "github.com/dtauraso/wirefold/Trace"
)

// --- node path construction ---
//
// A node owns every path under its own <root>/nodes/<id>/ directory EXCEPT
// nodes/<id>/edges/ (that subtree belongs to the edgeMover of each edge leaving this
// node — see edge_mover.go's doc comment and .claude/rules/persistence-ownership.md
// "The model"). These are the ONLY functions in the package that build those paths;
// quant_offset_persist.go and scene_anchor_persist.go call them rather than
// constructing the path themselves. safeTreePathComponent (scene_persist.go) is applied
// at each call site before use, same as it always was — node ids and port names are
// spec-authored and must not escape the tree root via ".." or a separator.

// positionFilePath is <root>/nodes/<id>/position.json — a node's exact scene-polar
// position plus its quantized-scalar-triple cache (quant_offset_persist.go).
func positionFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "position.json")
}

// pendingSend is one (destination, message) pair this node's own goroutine tried to
// deliver, failed (the target's inbox was momentarily full), and is retrying — see
// nodeMover.pending's doc comment. There is no separate sender goroutine:
// only nm's own goroutine ever reads or writes nm.pending.
type pendingSend struct {
	destID string
	msg    moveMsg
}

// nodeMover owns one node's geometry. It drains its own dedicated inbound channels
// (extIn + one per neighbor — there is no single shared inbox) in its own goroutine
// and, on a move for itself, updates its held position and re-emits its node-geometry.
type nodeMover struct {
	id   string
	geom nodeGeom
	// persistRoot is the tree root this node's mover writes its OWN per-node files
	// (position.json — quant_offset_persist.go; port anchor files —
	// scene_anchor_persist.go) into. Set once, for every nodeMover, by
	// MoveDispatch.EnableEditPersist after the startup seed (mirrors md.persist's other
	// armed-after-seed fields). Empty ("") means unarmed — bare test construction, or a
	// MoveDispatch built without EnableEditPersist — and every persist* method below is a
	// no-op. This node's own goroutine reads it only from its own persist* methods, so no
	// synchronization is needed even though every mover shares the same EnableEditPersist
	// call that sets it (a plain string write before any mover goroutine starts, same
	// happens-before shape as clockSrc/speedCh below).
	persistRoot string
	// extIn is this node's dedicated channel for EXTERNAL entries — the stdin/gesture
	// goroutine's drag/dragStart sends (md.sendMove). Nothing else ever writes here: no
	// other mover shares it.
	extIn chan moveMsg
	// neighborIn holds one dedicated inbound channel PER ADJACENT NODE (keyed by that
	// neighbor's id) — the "two channels, A→B and B→A" topology generalized to this
	// node's whole neighbor set. Built once at construction (newMoveDispatch) from edge
	// adjacency and never mutated afterward, so it's safe for run() to snapshot into a
	// fixed select-case list at goroutine start. A neighbor M's own goroutine is the only
	// writer of neighborIn[M]; nothing else ever sends on it.
	neighborIn map[string]chan moveMsg
	tr         *T.Trace
	// clockSrc is the Clock this nodeMover's own goroutine (run) Copies from EXACTLY
	// ONCE at its own start, into
	// clk below — the same pattern edgeMover.run and DriveHeld already use, so the
	// mover is no longer the odd one out pacing on a bare wall-clock timer. Not read
	// again after that copy.
	clockSrc wire.Clock
	// clk is this nodeMover's OWN clock copy, set once by run() at goroutine start.
	// Only this goroutine ever reads it. Defaults to a fresh, real, live-ticking
	// RealClock (see newNodeMover) so a test that never launches run() (e.g. a bare
	// nodeMover literal driving flushPending directly) never dereferences a nil Clock.
	clk wire.Clock
	// speedCh delivers a speed change to THIS nodeMover's own clk copy
	// (per-goroutine-clock.md "Delivery"), polled via ApplySpeedNonBlocking every
	// cycle of run's loop. Set once, at construction (newMoveDispatch), from the
	// loader's build-wide speed-sink accumulator; nil in bare test construction, which
	// is fine — ApplySpeedNonBlocking is a no-op on a nil channel.
	speedCh chan float64
	// There is no geomMu. m.geom (port_geometry.go) splits into an embedded, write-once
	// nodeIdentity (Kind/Label/R/SceneCenter — set once at construction in loader.go,
	// grepped clean of any later write anywhere in this package) and MUTABLE state
	// (ScenePolar/HasPos/ReachR) written only by applyCenter. Every writer AND every
	// reader of the mutable part — applyCenter, emitGeometry's
	// full-struct copy — runs exclusively on nodeMover's OWN inbox-drain goroutine
	// (run/handle), so there is never more than one goroutine touching that memory. The
	// one cross-goroutine reader, MoveDispatch.NodeKind (node_move.go), called from the
	// gesture/stdin-reader goroutine, reads ONLY nm.geom.Kind — a field on the embedded
	// nodeIdentity, which no writer here ever touches. So the property that would
	// require synchronization (a mutable field read cross-goroutine, or an identity field
	// that could gain a second writer) provably doesn't hold, by construction of the type
	// split, not by coincidence of which byte ranges happen to overlap today.
	//
	// CHECKED BY CODE: TestNodeKindConcurrentWithApplyCenterUnderRace
	// (node_mover_geom_race_test.go) drives NodeKind's reader loop and applyCenter's
	// writer loop concurrently under -race, as a standing
	// regression check that the split holds (a future change reintroducing a write to an
	// identity field, or widening NodeKind's read to a whole-struct copy, would make it
	// fail). There is no separate per-node "Update()" writer goroutine — that was the
	// retired SLICE 3 architecture.
	// centerOut is this node's OWN dedicated one-slot delivery channel to the
	// DISPATCH goroutine's owned center mirror (moverRegistry.centerMirror).
	// A size-1 buffered channel written with LATEST-WINS semantics (applyCenter drains
	// any stale unread value before sending the fresh one, never blocking): only the
	// newest pushed center matters to a framing read, so an unread stale value is
	// simply overwritten rather than queued. Only this node's own goroutine
	// (applyCenter) ever sends here; only the dispatch goroutine (moverRegistry.
	// drainCenterMirror) ever receives.
	centerOut chan vec3
	// sendMove routes a moveMsg to another id's OWN dedicated channel (resolveDest, above)
	// — no shared inbox, no shared mutable state.
	// Bound to md.enqueueFor(nm): it appends to nm.pending and immediately attempts a
	// non-blocking flush (never blocks the calling handler goroutine).
	sendMove func(id string, msg moveMsg)
	edgeIDs  []string
	// centerOf resolves another node's current world center, bound to
	// md.centerOfNode. Unused by any live handler now that the rule/gate/anchor
	// cascade (which used it to read rule-neighbor centers) is gone; kept wired for
	// any future direct-neighbor lookup need.
	centerOf func(id string) (vec3, bool)
	// commitLocal is the OWNER-GOROUTINE commit path, bound to
	// md.commitNodeMoveLocal (generalized to every node). It publishes this node's
	// own snap SYNCHRONOUSLY via applyCenter instead of enqueuing an async self-send,
	// so it is safe to call from THIS node's own handle() for a moveMsgKindDrag, with
	// no cross-goroutine self-send and no shared mutable state (each node's quantized
	// offset lives on its own mover — see nodeMover.quantOffset). nil in tests that
	// build a bare nodeMover directly.
	commitLocal func(id string, newPos vec3)
	// partnerCenters is THIS node's OWN copy of every direct neighbor's last-known world
	// center, read by quantized_move.go's neighbor-move math (neighborSetCReposition et
	// al.). Written ONLY by this node's own goroutine: seeded once at construction
	// (newMoveDispatch, single-threaded setup) from each neighbor's load-time geom, then
	// kept current by the moveMsgKindNeighborCenter handler in handle() below, fed by
	// every direct neighbor's own applyCenter push. Never read or written by any other
	// goroutine.
	partnerCenters map[string]vec3
	// quantOffset is THIS node's own quantized polar offset (iTheta,iPhi,iR + step
	// constants) about the scene center — the per-node replacement for the formerly
	// shared md.quantizedOffsets map, which one mover goroutine's read could race
	// another mover goroutine's write on the SAME Go map object even for different
	// keys (fatal "concurrent map read and map write"). Seeded at load
	// (buildMoveDispatch) from the computed/persisted offset, then mutated ONLY by
	// this node's own commit path (commitNodeMoveCommon, called from this node's own
	// goroutine via commitLocal) — single-writer, no map, no race.
	quantOffset quantizedOffset
	// pending is THIS node's own outbound retry queue: sendMove appends here and attempts an immediate
	// non-blocking send; an item that can't be delivered right now (the target's
	// inbox is momentarily full) stays here and is retried — before any newer item to
	// the SAME destination — on the next flushPending call, which nm's own run loop
	// makes every cycle. There is no dedicated sender goroutine: only
	// nm's own goroutine ever touches nm.pending (every sendMove call originates from
	// nm.handle, which only ever runs on nm's own run-loop goroutine). This is the
	// same retain-and-retry shape PacedWire already uses for its outCh delivery
	// handoff (full → retry next cycle, bead stays in inflight) — reused rather than
	// a second invented pattern.
	pending []pendingSend
	// tap is a TEST-ONLY observability seam: when non-nil, THIS mover's own enqueueFor
	// closure invokes it with every (destID, msg) it routes, before appending to
	// pending. nil in production — production code never calls MoveDispatch.SetMsgTap,
	// so this stays nil and every enqueueFor call skips it with one plain nil check.
	// Owned entirely by this mover: set once before Start (by
	// SetMsgTap, which runs on the setup goroutine before any mover goroutine is
	// launched — happens-before every later read) and read only by this mover's own
	// enqueueFor closure, which only ever runs on this mover's own goroutine. It is pure
	// observation — it never authors domain state or changes routing.
	tap func(destID string, msg moveMsg)
	// resolveDest looks up the ONE dedicated directed channel FROM this node TO the
	// given destination id — the destination's neighborIn[this node's id] if destID is
	// another node, or the destination edge's srcIn/dstIn depending on which endpoint
	// this node is (md.mr.nodeMovers/md.mr.edgeMovers are read-only directories after
	// construction, safe to read from any goroutine). There is no shared inbox to look
	// up: every (sender, destination) pair resolves to its OWN channel. nil only in
	// tests that build a bare nodeMover directly, in which case flushPending is a no-op.
	resolveDest func(id string) (chan moveMsg, bool)

	// --- chain bead actors (bead_chain.go, PLAN.md "two clocks per bead") ---
	// beadTickFn, when non-nil, is this node's production hook for a fresh dedicated
	// human-clock subscription (wire.NewTickChan) handed to each newly-started chain
	// bead goroutine. nil in every bare-literal test nodeMover (chain_beads_test.go and
	// friends) — that absence is what keeps chainBeads a pure, synchronous function with
	// no live TickBroadcaster goroutine in those tests; set once, in newNodeMover, for
	// every production node.
	beadTickFn func() <-chan struct{}
	// beadChains holds this node's own live per-outgoing-edge bead-actor chain, keyed by
	// target id. Written and read ONLY by this node's own goroutine (chainBeads, called
	// from writeStreamFrame/run — never a second goroutine). nil until the first
	// production reconcile (see reconcileBeadChain).
	beadChains map[string]*edgeBeadChain

	// --- dedicated per-node stream (memory/feedback_no_single_writer_bridge.md) ---
	// streamOut, when Ok(), is THIS node's OWN dedicated fd (see
	// MoveDispatch.SetNodeStreams / Buffer/stream_fds.go's StreamKindNode). A dead
	// claimedStream (the default — no WIREFOLD_STREAM_FDS "node" entry, e.g. headless
	// tests, OR a rejected second claim — see stream_claim.go) means writeStreamFrame is
	// a no-op: this node's geometry+ports+label are simply never written to a per-node
	// stream. claimedStream's unexported field + unexported constructor
	// (newClaimedStream, called only from setNodeStreams) make a SECOND claim on this
	// node's fd structurally rejected, not just written ONLY by this nodeMover's own
	// goroutine (emitGeometry/run) by convention.
	streamOut claimedStream
	// nodeRow is this node's stable buffer NODE-ROW index (the seed order — see
	// MoveDispatch.SetNodeStreams), carried on every Port row this node's stream frame
	// writes so a port row can be resolved back to (nodeRow, portIndex) on the TS side
	// without a shared port table.
	nodeRow int32
	// selfKind is this node's own kind name (specNode.Type), set ONCE at construction
	// (build.go's buildMoveDispatch) alongside neighborKinds.
	selfKind string
	// outTargets is THIS node's own OUTGOING edge targets (b.spec.Edges where Source ==
	// this node), seeded once at load beside neighborKinds (build.go) and never written
	// again — the set of chains this node owns, matching where the edge is stored on disk
	// (topology/nodes/<source>/edges/, outgoing only). Read only by chainBeads on this
	// node's own goroutine.
	outTargets []string
	// outWires / outWireTargets are THIS node's own outgoing wires and the target id each
	// one goes to, parallel slices bound once at load (moverRegistry.bind) and never
	// written again. This node's own goroutine DRIVES these wires (run, below) — there is
	// no wire goroutine — and reads their in-flight fractions to light its own chain.
	//
	// outWireTargets is separate from outTargets because a spec edge may have no bound
	// wire (an unresolved handle leaves slotReg without an entry), so the two are not
	// index-parallel. Lighting matches by target id, not by position.
	outWires       []*wire.PacedWire
	outWireTargets []string
	// outWireOuts is the *wire.Out for each entry in outWires (parallel, same index),
	// bound alongside it in moverRegistry.bind. This node's own chainBeads (this node
	// IS the count owner, docs/bead-lattice.md "Ownership") calls PublishSteps on each
	// entry every pass, so the wire's own timing budget always reads the same integer
	// the chain was just laid out on. nil when this edge's source handle wasn't found
	// in outSink (chainBeads then just skips the publish for that edge).
	outWireOuts []*wire.Out
	// outStepsIn is outWireOuts' sibling: each entry's edgeMover.stepsIn channel
	// (edge_mover.go's doc comment), parallel to outWires/outWireOuts by index.
	// chainBeads sends the same freshly computed step count here as it publishes to
	// outWireOuts, so the edgeMover's own goroutine — which cannot read the Out's
	// Geom() cache directly without racing its one owning goroutine — can revise an
	// in-flight bead's remaining travel (ReviseInFlightGeometry) against the current
	// count too. nil entries are skipped by sendStepsNonBlocking (a nil channel's send
	// case is simply never selected).
	outStepsIn []chan int
	// neighborKinds maps each DIRECT domain-adjacent neighbor id → that neighbor's kind
	// name, derived once from the loaded spec's node list + edges (build.go's
	// buildMoveDispatch) — never touched again. Indexing a nil/missing entry is safe
	// (returns ""), so no init needed. Used for neighbor-kind-dependent geometry
	// (edgeStepCount, nodeTorusOuterR) without a central id->kind table.
	neighborKinds map[string]string
	// mutualTargets marks each outgoing target that ALSO has an edge back to this node.
	// Such a pair's two chains run along the same centre line and would draw on top of
	// each other, so each end offsets its own chain perpendicular to that line
	// (parallelChainOffset, port_geometry.go). Seeded once at construction from the loaded
	// edge set, on the single-threaded setup path, and read only by this node's own
	// goroutine afterwards — a load-time fact, not shared state.
	mutualTargets map[string]bool
	// coplanarEdges: this node's ring plane must CONTAIN the edge leaving it, so the chain
	// and both tori share one plane (scene_tabs.go's CoplanarEdges). Set once at
	// construction from the loaded scene; read only by this node's own goroutine.
	coplanarEdges bool
	// upAxis: this node's ring axis and its own drawn vector both point at world +y
	// (scene_tabs.go's UpAxis). Set once at construction from the loaded scene.
	upAxis bool
	// nodeRowFor resolves a node id to its buffer NODE-ROW index (mirroring the old
	// central accumulator's NodeRowFor), injected via MoveDispatch.SetNodeStreams so this
	// package stays Buffer-independent. Used to resolve a neighbor id to its row for
	// breadcrumb TargetRow columns (handle's moveMsgKindNeighborCenter case) and chain-aim
	// diagnostics (chain_beads.go).
	nodeRowFor func(id string) (int32, bool)
	// --- own selection/hover/abc-drag UI state (per-owner, no shared/republished map) ---
	//
	// This node's OWN current selected/hovered/latchedSel bits — set only by THIS
	// node's own goroutine, from messages the gesture goroutine sends on extIn
	// (moveMsgKindSelect/Hover/Latched). No lock: only nm.handle (this goroutine) ever
	// writes these, and writeStreamFrame (also this goroutine) is the only reader.
	selected, hovered, latchedSel uint8
	// hoverPort/hoverIsInput name the specific port currently hovered on this node (""
	// = whole-node hover, only meaningful when hovered==1). Set alongside hovered by a
	// moveMsgKindHover message.
	hoverPort    string
	hoverIsInput bool
	// kindID is this node's static numeric kind (Buffer.NodeKindID) — set ONCE at
	// construction (MoveDispatch.SetNodeStreams), never touched again: a node's kind
	// never changes after load, so there is no per-emit lookup to perform.
	kindID uint8
	// buildFrame packs this node's combined per-fd frame (node fields + ports + label)
	// using Buffer's own row-writer columns (Buffer.BuildNodeStreamFrame), injected so
	// this package needs no Buffer import.
	buildFrame func(tick uint32, nodeRow int32, nodeID int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, poleTheta, polePhi, ringAxisTheta, ringAxisPhi, tiltVectorLen, tiltVectorTheta, tiltVectorPhi, coplanarNormalTheta, coplanarNormalPhi float32, selected, kindID, hovered, latchedSel uint8, label string, chainBeadOX, chainBeadOY, chainBeadOZ []float32, chainBeadLit []uint8, chainBeadLitValue []int32, events []wire.RowEvent) []byte
	// tiltVectorThetaIdx/tiltVectorPhiIdx are THIS node's own vector direction, as INTEGER
	// indices into TiltVectorAngleStep (memory/feedback_abc_times_constant_not_rederive.md
	// — index × step-constant, trig only at the cartesian/polar boundary). Default 0,0
	// means world +y (θ=0), matching the pre-existing hardcoded +y direction. Always
	// persisted by this node's own mover into ITS OWN position.json (quant_offset_persist.go)
	// and streamed by writeStreamFrame below, but who DECIDES the value differs by kind:
	// for a kind that claims BuildArgs.TiltEditIn and owns its own index independently
	// (Node1/Node2 today), this mover is a passive MIRROR — written only by
	// moveMsgKindTiltIndexSync, a one-way notification from that node's own goroutine,
	// never decided or mutated here. For every other kind, this mover remains the sole
	// decider/mutator, written directly by an edit-update(tiltVector) message
	// (moveMsgKindTiltVectorAngle, applyUpdateTiltVector's fallback) or seeded once from
	// the persisted load value (build.go).
	tiltVectorThetaIdx, tiltVectorPhiIdx int32
}

func newNodeMover(id string, geom nodeGeom, tr *T.Trace, clockSrc wire.Clock) *nodeMover {
	// clk defaults to a fresh RealClock (its own independent origin — fine here: this
	// default is only ever read by a test that never launches run() as a goroutine;
	// production always overwrites it below with clockSrc.Copy() before the goroutine
	// does anything else), matching newEdgeMover's same default for the same reason.
	nm := &nodeMover{
		id: id, geom: geom,
		extIn: make(chan moveMsg, moverInboxDepth), neighborIn: map[string]chan moveMsg{}, tr: tr,
		partnerCenters: map[string]vec3{},
		centerOut:      make(chan vec3, 1),
		clockSrc:       clockSrc, clk: wire.NewRealClock(),
	}
	// Self-seed centerOut with the initial geometry (even when !HasPos, in which case
	// nodeWorldPos falls back to the origin) so the dispatch goroutine's first drain
	// always finds a valid center — covers every construction path (not just
	// newMoveDispatch's loop, which additionally seeds moverRegistry.centerMirror
	// directly before any mover goroutine runs).
	nm.centerOut <- nodeWorldPos(geom)
	// Production-only hook: arms the bead-actor path in chainBeads/reconcileBeadChain
	// (bead_chain.go). Bare `&nodeMover{...}` test literals never call newNodeMover, so
	// beadTickFn stays nil there and chainBeads' pure-function tests never touch a live
	// TickBroadcaster goroutine — see beadTickFn's own doc comment.
	nm.beadTickFn = wire.NewTickChan
	return nm
}

// handle applies one move to this node: update held position, re-emit node-geometry.
func (m *nodeMover) handle(msg moveMsg) {
	if msg.NodeID != m.id {
		return
	}
	if msg.Kind == moveMsgKindCenter {
		// nodeMover is the SOLE writer of its own position (single-writer by
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
		// offset is ever touched by a foreign mover goroutine): commit this node's OWN
		// new position via the local (synchronous-snap-publish) commit path. A drag is
		// always a FREE move now -- there is no equal-radii solve and no propagation
		// past this node's own commit.
		newPos := msg.Target
		if m.commitLocal != nil {
			m.commitLocal(m.id, newPos)
		}
		if m.tr != nil {
			m.tr.Breadcrumb("drag.commit", m.id, "", fmt.Sprintf("newPos=(%.4f,%.4f,%.4f)", newPos.X, newPos.Y, newPos.Z))
			// Structured buffer counterpart, riding this node's own dedicated
			// stream frame (emitGeometry's own next emit already fires from
			// commitLocal above, so this rides as a distinct events-only-shaped
			// write here rather than waiting on that one).
			m.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbDragCommit, Debug: 1,
				NodeRow: m.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
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
			m.selected = 1
		} else {
			m.selected = 0
		}
		return
	}
	if msg.Kind == moveMsgKindHover {
		if msg.Bool {
			m.hovered = 1
			m.hoverPort = msg.Port
			m.hoverIsInput = msg.IsInput
		} else {
			m.hovered = 0
			m.hoverPort = ""
			m.hoverIsInput = false
		}
		return
	}
	if msg.Kind == moveMsgKindLatched {
		if msg.Bool {
			m.latchedSel = 1
		} else {
			m.latchedSel = 0
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
		if msg.Axis == "phi" {
			m.tiltVectorPhiIdx += delta
		} else {
			m.tiltVectorThetaIdx += delta
		}
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: this path only runs for a kind that has NOT claimed BuildArgs.TiltEditIn
		// (every kind except Node1/Node2 today — see moveMsgKindTiltVectorAngle's own doc
		// comment and applyUpdateTiltVector's fallback, stdin_reader.go). Node1/Node2's own
		// tilt-panel edits are routed to their OWN goroutine instead (TiltEditIn), which
		// applies the click, syncs this value back via moveMsgKindTiltIndexSync, AND places
		// "the kick" bead on its own Out directly — none of that happens here anymore.
		return
	}
	if msg.Kind == moveMsgKindTiltVectorReset {
		// Return THIS node's own vector direction to the start position — both indices to
		// 0, the documented default (tilt vector at world +y). No bead: this is a
		// stop-and-return, not a kick. Persisted immediately, same as an adjust.
		m.tiltVectorThetaIdx = 0
		m.tiltVectorPhiIdx = 0
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: same split as moveMsgKindTiltVectorAngle — this path only runs for a kind
		// that has NOT claimed BuildArgs.TiltEditIn. Node1/Node2 route a reset through
		// their own TiltEditIn/TiltEditMsg.Reset instead.
		return
	}
	if msg.Kind == moveMsgKindTiltIndexSync {
		// Passive mirror only: Node1/Node2's own goroutine already decided and mutated
		// its OWN index (reactToArrival/panel-edit handling now live there —
		// nodes/Node1/node.go, nodes/Node2/node.go). This mover just applies exactly what
		// it is told, persists it to this node's OWN position.json, and re-emits so the
		// panel's read-only reflect and the drawn arrow both pick up the change.
		m.tiltVectorThetaIdx = msg.ThetaIdx
		m.tiltVectorPhiIdx = msg.PhiIdx
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		return
	}
	if msg.Kind == moveMsgKindNeighborCenter {
		// Delivery-mechanism push (see applyCenter/partnerCenters' doc comments): a
		// direct neighbor's OWN center just changed. Store it in THIS node's owned
		// partnerCenters map (write, own goroutine only) and re-emit THIS node's own
		// geometry so its aimed ports pick up the fresh partner center — same value,
		// same effect as the old cross-goroutine snap read, just message-delivered.
		// ONE HOP ONLY: this node's own center did NOT change, so it must never push
		// a NeighborCenter of its own onward from here (no cascade past this point).
		if m.partnerCenters == nil {
			m.partnerCenters = map[string]vec3{}
		}
		m.partnerCenters[msg.SenderID] = msg.FromCenter
		if m.tr != nil {
			// DIAGNOSTIC ONLY (task/log-node4-chain-aim): records that this node's own
			// goroutine received a neighbor-center push, and from whom, so a drag-time
			// trace can show whether/when it arrives relative to this node's own emits.
			value := fmt.Sprintf("sender=%s center=(%.4f,%.4f,%.4f)", msg.SenderID, msg.FromCenter.X, msg.FromCenter.Y, msg.FromCenter.Z)
			m.tr.Breadcrumb("neighbor-center-recv", m.id, msg.SenderID, value)
			senderRow := int32(-1)
			if m.nodeRowFor != nil {
				if r, ok := m.nodeRowFor(msg.SenderID); ok {
					senderRow = r
				}
			}
			m.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbNeighborCenterRecv, Debug: 1,
				NodeRow: m.nodeRow, PortRow: -1, TargetRow: senderRow, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
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

// PerpendicularThetaIdx is the tiltVectorThetaIdx value at which the tilt vector is exactly
// perpendicular to world +y: CurveParamTiltVectorAngleStep is π/12 (15°), and π/2 (90°) is
// exactly 6 steps. Comparing to this INTEGER is what makes the straightening loop's stop
// condition exact — cos(π/2) in float64 is ~6.1e-17, so a literal float dot==0 test would
// never fire (memory/feedback_abc_times_constant_not_rederive.md: index arithmetic, trig
// only at the cartesian/polar boundary). Exported (capitalized) so Node1/Node2's own
// goroutine — which now runs the straightening rule itself, per-package — can compare
// against it without duplicating the constant; the rule itself no longer lives here (see
// nodes/Node1/node.go, nodes/Node2/node.go).
//
// dot(tilt, coplanarNormal) == 0 is decided as thetaIdx == PerpendicularThetaIdx, not by
// computing an actual float dot product. STATE THE ASSUMPTION THAT MAKES THE SHORTCUT
// VALID: the tilt vector's in-plane angle IS its θ index only because, for this scene, the
// ring plane holds world +y and θ is measured from +y, so the two coincide (see
// tiltVectorThetaIdx's own doc comment and the CoplanarNormal/UpAxis derivations in
// writeStreamFrame above). A scene whose ring plane does NOT contain +y breaks that
// coincidence — θ would then measure something unrelated to the coplanar normal, and the
// rule would need to compare an actual dot(tilt, coplanarNormal) via the two integer
// indices' angles converted through anglesToWorldOffset, not thetaIdx alone.
const PerpendicularThetaIdx int32 = 6

// applyCenter is the SOLE WRITE of this node's center/reach. It is called ONLY from
// this nodeMover's own inbox-drain goroutine (handle's moveMsgKindCenter case, driven
// by fanCenters below), which is what makes that one goroutine the exclusive writer of
// m.geom. It sets the held polar position, pushes the fresh center to the dispatch
// goroutine's owned center mirror (m.centerOut, latest-wins — see its doc comment) and
// to every direct neighbor's partnerCenters map (below), and re-emits this node's live
// geometry.
func (m *nodeMover) applyCenter(center vec3, reach float64) {
	setNodeWorld(&m.geom, center)
	m.geom.ReachR = reach
	// Latest-wins non-blocking push onto centerOut: drain any stale unread value first
	// so the slot always ends up holding the newest center, never blocking this
	// goroutine even if the dispatch goroutine hasn't drained the previous push yet.
	select {
	case <-m.centerOut:
	default:
	}
	select {
	case m.centerOut <- center:
	default:
	}
	// Push this fresh center to every direct neighbor (nm.neighborIn's key set — one
	// hop, no cascade) so each neighbor's OWN partnerCenters map picks it up via
	// moveMsgKindNeighborCenter (handle, below). Routed through m.sendMove (this
	// node's own retry queue), same as every other fan-out this file makes, so a
	// momentarily-full neighbor inbox is retried, never dropped or blocking. Sent
	// BEFORE this same commit's broadcastToEdgesAndPartners nil-Center re-emit (called
	// right after applyCenter by every live caller), so per-destination FIFO delivers
	// this push first and the re-emit always sees the just-pushed center.
	for neighborID := range m.neighborIn {
		m.sendMove(neighborID, moveMsg{Kind: moveMsgKindNeighborCenter, NodeID: neighborID,
			SenderID: m.id, FromCenter: center})
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

// emitGeometry re-emits this node's authoritative geometry (center, radius, ring
// normals — no port geometry: a port carries none, docs/channels-not-ports.md).
// This method and applyCenter both run on nodeMover's own inbox-drain goroutine only
// (see the doc comment on nodeMover.geom), so a plain field read here can never race a
// concurrent writer.
func (m *nodeMover) emitGeometry() {
	// Dedicated per-node stream (see streamOut's doc comment): write this node's own
	// combined frame immediately on a geometry change, in addition to the tick-driven
	// write in run()'s loop (mirrors edgeMover.recomputeGeometry's writeStreamFrame call).
	// NodeGeometry rides THIS frame's own EVENTS section (fully decentralized — it
	// never rides the VIEW stream's fallback bucket) — this
	// nodeMover is the sole owner of its node's geometry, so it resolves its own
	// NodeRow at the call site (owner_events.go) rather than routing through a
	// shared accumulator.
	m.writeStreamFrame([]wire.RowEvent{{
		Kind: T.KindNodeGeometry, NodeRow: m.nodeRow,
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}})
}

// writeStreamFrame packs and writes this node's combined per-fd frame (center/radius/
// ring-normals + ports + label + selection-UI columns) to its OWN dedicated fd
// (streamOut). No-op when streamOut is nil (the fallback — see its doc comment) or
// buildFrame was never injected (bare test construction). Called only by this nodeMover's
// own goroutine (emitGeometry and run's per-cycle loop), reading
// m.geom. events carries whatever this call's caller wants riding this frame's trailing
// EVENTS section (nil from run()'s plain tick-driven write).
func (m *nodeMover) writeStreamFrame(events []wire.RowEvent) {
	if !m.streamOut.Ok() || m.buildFrame == nil {
		return
	}
	// INVARIANT: a mover carries only its OWN node's events on its OWN dedicated stream.
	// This is the per-goroutine bridge stated in CLAUDE.md's "Bridge surface" and in
	// memory/feedback_no_single_writer_bridge.md + memory/feedback_per_goroutine_bridge.md,
	// and until now it was enforced by prose alone. NodeRow is the ownership column; a
	// FOREIGN node is referenced through TargetRow (see quantized_move.go's abc-drag
	// breadcrumb, which sets NodeRow: nm.nodeRow and TargetRow: the other node). Violating
	// it produces a frame the TS side decodes onto the wrong row — a silently wrong scene
	// that still renders, which is the expensive failure this panic converts into a cheap
	// one. Placed AFTER the nil guard on purpose: bare movers built in tests never reach
	// the pack path, and nodeRow is seeded alongside streamOut (stream_wiring.go), so any
	// frame that gets here has a real row.
	for _, e := range events {
		if e.NodeRow != m.nodeRow {
			panic(fmt.Sprintf(
				"nodeMover.writeStreamFrame: node %q (row %d) is carrying a %s event for row %d on its OWN dedicated stream — NodeRow is an ownership claim, not a reference; a foreign node belongs in TargetRow",
				m.id, m.nodeRow, e.Kind, e.NodeRow))
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
	// tiltVectorLen is this node's own drawn vector, along the SAME axis as its ring, and 0
	// where a scene draws none (Buffer/layout.go's TiltVectorLen). It runs from the node's
	// centre to its own top, so its length IS the node's radius.
	var tiltVectorLen float64
	if m.upAxis && m.geom.HasPos && len(m.partnerCenters) == 1 {
		// UPRIGHT: the ring STANDS UP along its edge — its plane holds both the edge and
		// world +y, so the node's own up-vector lies IN the ring's plane rather than
		// sticking out of a flat disc. An axis of +y itself would lie the ring flat and
		// put the vector perpendicular to it, which is the opposite arrangement.
		for _, partner := range m.partnerCenters {
			if t, p, ok := uprightRingAxis(nodeWorldPos(m.geom), partner); ok {
				ringAxisTheta, ringAxisPhi = t, p
			}
		}
		tiltVectorLen = nodeRadius(m.geom.Kind)
	} else if m.coplanarEdges && m.geom.HasPos && len(m.partnerCenters) == 1 {
		// COPLANAR EDGES: swing the axis off the inward pole by the smallest amount that
		// puts the edge INSIDE the ring plane — the inward pole with its along-the-edge
		// component removed. The chain, this node's torus and the beads' own tori then
		// share one plane instead of the chain running through the holes. Only for a node
		// with exactly ONE neighbour: two non-collinear edges have no common plane.
		for _, partner := range m.partnerCenters {
			if t, p, ok := poleContainingEdge(poleTheta, polePhi, nodeWorldPos(m.geom), partner); ok {
				ringAxisTheta, ringAxisPhi = t, p
			}
		}
	}
	// tiltVectorTheta/tiltVectorPhi are this node's OWN vector direction — separate from the ring
	// axis above, so a scene/user can aim a node's vector somewhere other than its ring.
	// Never a free float: index × TiltVectorAngleStep (see the constant's own doc comment),
	// the streamed value is pure arithmetic on the integer state this node's own mover
	// holds and persists (m.tiltVectorThetaIdx/tiltVectorPhiIdx).
	tiltVectorTheta := float64(m.tiltVectorThetaIdx) * CurveParamTiltVectorAngleStep
	tiltVectorPhi := float64(m.tiltVectorPhiIdx) * CurveParamTiltVectorAngleStep
	// The COPLANAR NORMAL: the in-plane direction toward this node's partner
	// (coplanarNormalTowardPartner, port_geometry.go) — derived from the EDGE, never from
	// the tilt vector, so turning the tilt never moves what it is measured against (see
	// that helper's doc comment). Zero when this node draws no vector at all, matching the
	// first.
	var coplanarNormalTheta, coplanarNormalPhi float64
	if tiltVectorLen > 0 {
		for _, partner := range m.partnerCenters {
			if t, p, ok := coplanarNormalTowardPartner(nodeWorldPos(m.geom), partner, ringAxisTheta, ringAxisPhi); ok {
				coplanarNormalTheta, coplanarNormalPhi = t, p
			}
		}
	}
	label := m.geom.Label
	if label == "" {
		label = m.id
	}
	selected, hovered, latchedSel, kindID := m.selected, m.hovered, m.latchedSel, m.kindID
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
	// persistence-ownership.md), so it is m.nodeRow+1 by construction — not re-derived by any
	// offline rule the decoder also has to apply, it travels with the frame.
	frame := m.buildFrame(uint32(m.clk.Tick()), m.nodeRow, m.nodeRow+1,
		float32(center.X), float32(center.Y), float32(center.Z),
		float32(nodeRadius(m.geom.Kind)), float32(sphereR),
		verticalRingNormalX, verticalRingNormalY, verticalRingNormalZ,
		flatRingNormalX, flatRingNormalY, flatRingNormalZ,
		float32(poleTheta), float32(polePhi), float32(ringAxisTheta), float32(ringAxisPhi), float32(tiltVectorLen),
		float32(tiltVectorTheta), float32(tiltVectorPhi), float32(coplanarNormalTheta), float32(coplanarNormalPhi),
		selected, kindID, hovered, latchedSel,
		label, chainOX, chainOY, chainOZ, chainLit, chainLitVal, events)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	// Fire-and-forget, same reasoning throughout this bridge: no delivery
	// guarantee on this channel, errors ignored.
	_, _ = m.streamOut.Write(hdr[:])
	_, _ = m.streamOut.Write(frame)
}

// flushPending retries every message in m.pending in order, attempting a non-blocking
// send to its destination's inbox. A destination whose channel is momentarily full
// stays in the queue (retried next call) — and so does every LATER item addressed to
// that SAME destination, even if its own channel isn't full, so per-destination FIFO
// is preserved (a retained item is never overtaken by a newer one to the same
// destination). An item whose destination doesn't resolve (unknown id) is dropped,
// matching the old deliverMove no-op for an unknown id. Called only from m's own
// goroutine (sendMove, at enqueue time, and run's own loop, every cycle).
func (m *nodeMover) flushPending() {
	if len(m.pending) == 0 || m.resolveDest == nil {
		return
	}
	blocked := map[string]bool{}
	kept := m.pending[:0]
	for _, item := range m.pending {
		if blocked[item.destID] {
			kept = append(kept, item)
			continue
		}
		ch, ok := m.resolveDest(item.destID)
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
	m.pending = kept
}

// run is the node's per-goroutine move loop. It paces itself on its OWN clock copy the
// same way every other loop in the system does (edgeMover.run, DriveHeld,
// emitRefillSlide): a Clock.Copy()
// taken once here at goroutine start, ApplySpeedNonBlocking polled once per cycle, and
// SleepCycle(ctx) as the pacing sleep at the bottom of the loop. It used to be the odd
// loop out, blocking on a reflect.Select over its whole channel set instead; that is
// gone.
//
// Each cycle FIRST drains every one of its OWN dedicated inbound channels (extIn + one
// per neighbor, see the type's doc comment) — there is no shared inbox to drain
// — non-blockingly and acts on
// whatever is there, repeating the drain pass until a full pass finds nothing left (so a
// backlog on any one channel is fully drained before the cycle paces, not throttled to
// "one message per channel per cycle"), THEN retries any pending sends, THEN sleeps one
// clock cycle. Nothing here ever blocks on a receive OR a send: an empty channel just
// falls through its `default`, exactly the "read non-blockingly at the top, act on what's
// there, pace on the clock" shape the design calls for. This does not busy-wait (the
// pacing sleep bounds every cycle to one clock tick regardless of whether the drain pass
// found anything), and does not throttle a real backlog (a full pass draining every
// channel to empty runs entirely within the cycle, before the sleep — a burst of incoming
// messages is drained as fast as it arrives, not capped at one per channel per tick).
func (m *nodeMover) run(ctx context.Context) {
	if m.clockSrc != nil {
		m.clk = m.clockSrc.Copy()
	}
	// ONE-TIME startup geometry emit, on THIS node's own mover goroutine — this is now
	// the sole per-owner source of a node's initial node-geometry event (replacing the
	// old node-Update-loop startup emit builders.go's EmitGeometry closure used to make;
	// that closure no longer calls tr.NodeGeometry — see its doc comment). m.tr is
	// non-nil in production (newNodeMover always receives one); bare test construction
	// with a nil tr just skips this, matching emitGeometry's own nil-guard elsewhere.
	if m.tr != nil {
		m.emitGeometry()
	}
	for {
		wire.ApplySpeedNonBlocking(m.clk, m.speedCh)
		// Drain every dedicated inbound channel non-blockingly, repeating until a
		// full pass yields nothing — this is the "drain to empty, don't throttle a
		// backlog" half of the shape.
		//
		// Drain-until-empty, transitively bounded by each channel's own declared
		// capacity (moverInboxDepth) -- no iteration cap; see
		// nodes/wire/paced_wire.go's drainPlacements doc comment for the full
		// reasoning shared by every drain-until-empty loop in this repo.
		for {
			progressed := false
			select {
			case <-ctx.Done():
				return
			case msg := <-m.extIn:
				m.handle(msg)
				if msg.testDone != nil {
					close(msg.testDone)
				}
				progressed = true
			default:
			}
			for _, ch := range m.neighborIn {
				select {
				case msg := <-ch:
					m.handle(msg)
					if msg.testDone != nil {
						close(msg.testDone)
					}
					progressed = true
				default:
				}
			}
			if !progressed {
				break
			}
		}
		// Drive THIS node's own outgoing wires — placement drain, position step, delivery
		// — on this node's own goroutine and its own clock reading. This is the work
		// edgeMover.run used to do for the wire; the wire has no goroutine of its own
		// (docs/beads-are-the-edge.md step 3). Driving it here is also what makes
		// LiveBeadFractions safe to read below: same goroutine, no shared state.
		//
		// Read the clock ONCE for this whole pass, not once per wire: a per-wire read
		// can straddle a tick boundary mid-loop, splitting one emission's beads across
		// two ticks even though they were placed microseconds apart.
		outTick := m.clk.Tick()
		for _, pw := range m.outWires {
			pw.DriveOneCycle(ctx, outTick)
		}
		// Retry any pending sends (nm.pending/flushPending) every cycle — a
		// destination that was full earlier may have drained since.
		m.flushPending()
		// Selection/hover/drag UI state may have changed even with no geometry change
		// this cycle (that state is centrally owned elsewhere, not this nodeMover's own
		// — see uiStateFor's doc comment) — write this node's dedicated stream frame
		// every cycle (no-op when streamOut is nil), mirroring
		// edgeMover.run's same every-cycle writeStreamFrame call.
		m.writeStreamFrame(nil)
		if err := m.clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}
