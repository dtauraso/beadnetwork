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
	"io"
	"sync/atomic"

	T "github.com/dtauraso/wirefold/Trace"
)

// pendingSend is one (destination, message) pair this node's own goroutine tried to
// deliver, failed (the target's inbox was momentarily full), and is retrying — see
// nodeMover.pending's doc comment. There is no separate sender goroutine:
// only nm's own goroutine ever reads or writes nm.pending.
type pendingSend struct {
	destID string
	msg    moveMsg
}

// setPortAnchorId sets the AnchorId on the named port within the given geom,
// clearing any free-direction Anchor so AnchorId takes highest priority (matching
// portDir's resolution order: AnchorId > Anchor > side/slot). Returns true if the
// port was found and mutated. The geom is mutated in place (its slice elements are
// addressable). Used by both movers to apply a snapped ring-anchor update.
func setPortAnchorId(g *nodeGeom, port string, isInput bool, anchorId int) bool {
	list := g.Outputs
	if isInput {
		list = g.Inputs
	}
	for i := range list {
		if list[i].Name == port {
			list[i].AnchorId = &anchorId
			return true
		}
	}
	return false
}

// nodeMover owns one node's geometry. It drains its own dedicated inbound channels
// (extIn + one per neighbor — there is no single shared inbox) in its own goroutine
// and, on a move for itself, updates its held position and re-emits its node-geometry.
type nodeMover struct {
	id   string
	geom nodeGeom
	// extIn is this node's dedicated channel for EXTERNAL entries — the stdin/gesture
	// goroutine's drag/dragStart/anchor sends (md.sendMove, gesture.go's
	// applyRingAnchor). Nothing else ever writes here: no other mover shares it.
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
	clockSrc Clock
	// clk is this nodeMover's OWN clock copy, set once by run() at goroutine start.
	// Only this goroutine ever reads it. Defaults to a fresh, real, live-ticking
	// RealClock (see newNodeMover) so a test that never launches run() (e.g. a bare
	// nodeMover literal driving flushPending directly) never dereferences a nil Clock.
	clk Clock
	// speedCh delivers a speed change to THIS nodeMover's own clk copy
	// (per-goroutine-clock.md "Delivery"), polled via ApplySpeedNonBlocking every
	// cycle of run's loop. Set once, at construction (newMoveDispatch), from the
	// loader's build-wide speed-sink accumulator; nil in bare test construction, which
	// is fine — ApplySpeedNonBlocking is a no-op on a nil channel.
	speedCh chan float64
	// There is no geomMu. m.geom (port_geometry.go) splits into an embedded, write-once
	// nodeIdentity (Kind/Label/R/SceneCenter — set once at construction in loader.go,
	// grepped clean of any later write anywhere in this package) and MUTABLE state
	// (ScenePolar/HasPos/ReachR/Inputs/Outputs-element-AnchorId) written only by
	// applyCenter and handle's moveMsgKindAnchor case. Every writer AND every reader of
	// the mutable part — applyCenter, setPortAnchorId (via handle), emitGeometry's
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
	// snap is an atomically-published immutable snapshot of this node's current
	// center+reachR. Written only by the mover's own goroutine after every center
	// update; read by any goroutine (stdin reader) to observe the current position
	// without crossing into the mover's live geom.
	snap atomic.Pointer[centerSnap]
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
	// partnerCenter resolves, per (port,isInput) on this node, the CURRENT world center of
	// the single partner node connected via one edge (aimed-port model, port_geometry.go
	// portWorldPosAimed / builders.go partnerCenterFn). Wired by newMoveDispatch from
	// b.edgeEndpoints + the OTHER nodeMover's atomic snap — a dynamic, always-current lookup
	// with no shared mutable state. nil only in tests that build a bare nodeMover directly.
	partnerCenter partnerCenterFn
	// partnerCenters is THIS node's OWN copy of every direct neighbor's last-known
	// world center — the delivery-mechanism replacement for the old cross-goroutine
	// `other.snap.Load().c` read: partnerCenter's closure (above) now reads this map
	// instead of reaching into another mover's atomic snap. Written ONLY by this
	// node's own goroutine: seeded once at construction (newMoveDispatch, single-
	// threaded setup) from each neighbor's load-time geom, then kept current by the
	// moveMsgKindNeighborCenter handler in handle() below, fed by every direct
	// neighbor's own applyCenter push. Never read or written by any other goroutine —
	// the atomic snap (nm.snap) remains the cross-goroutine-safe publication point for
	// every OTHER reader (stdin reader, the gesture/quantize oracle); this map exists
	// solely to serve THIS node's own partnerCenter lookups without touching that
	// atomic.
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
	// neighborSetC runs THIS node's own plain-neighbor set-c redraw (keep stored
	// bearing, write only the new c, reposition self) — bound to
	// md.neighborSetCReposition. Dispatched from moveMsgKindNeighborSetC so a domain
	// neighbor's holder AND world position are written only by that neighbor's OWN
	// goroutine. nil in tests that build a bare nodeMover directly.
	neighborSetC func(selfID, fromID string, selfCenter, fromCenter vec3, deltaA, deltaB, deltaC int)
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
	// so this stays nil and every enqueueFor call skips it with one plain nil check, no
	// atomic, no lock. Owned entirely by this mover: set once before Start (by
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
	// layoutHolderFn resolves THIS node's own LocalPolar holder (md.lq.layoutHolders[id])
	// at CALL TIME rather than caching the *LayoutHolder at nodeMover construction:
	// buildMoveDispatch (which constructs nodeMovers) runs BEFORE buildNodes (which is
	// what actually populates md.lq.layoutHolders[id] on each node's embedded
	// LayoutHolder), so a value cached at construction would be permanently nil. The
	// map itself is a read-only directory once the whole load completes (same pattern
	// as dispatch/edgeIDs) — safe to read from any goroutine after that point. Read
	// here only by armDragAnchor, which runs exclusively on this node's own goroutine
	// (moveMsgKindDragStart, dispatched via handle).
	layoutHolderFn func() *LayoutHolder
	// dragAnchorByTo, dragAnchorArmed: THIS node's drag-anchor snapshot (see
	// moveMsgKindDragStart's doc comment) — the per-neighbor LocalPolar triples as of
	// the start of the CURRENT drag. Written only by armDragAnchor (moveMsgKindDragStart
	// handler) or by requantizeLocalPolars' lazy-arm fallback (first commit of a drag
	// that never got an explicit dragStart, e.g. a programmatic RootMove in a test) —
	// both run on this node's own goroutine. Cleared (dragAnchorArmed=false) by
	// armDragAnchor so a NEW drag on the same node always re-arms rather than reusing a
	// stale anchor from a previous drag.
	dragAnchorByTo  map[string]LocalPolar
	dragAnchorArmed bool

	// --- dedicated per-node stream (memory/feedback_no_single_writer_bridge.md) ---
	// streamOut, when non-nil, is THIS node's OWN dedicated fd (see
	// MoveDispatch.SetNodeStreams / Buffer/stream_fds.go's StreamKindNode). Nil (the
	// default — no WIREFOLD_STREAM_FDS "node" entry, e.g. headless tests) means
	// writeStreamFrame is a no-op: this node's geometry+ports+label are simply never
	// written to a per-node stream. Written ONLY by this
	// nodeMover's own goroutine (emitGeometry/run).
	streamOut io.Writer
	// nodeRow is this node's stable buffer NODE-ROW index (the seed order — see
	// MoveDispatch.SetNodeStreams), carried on every Port row this node's stream frame
	// writes so a port row can be resolved back to (nodeRow, portIndex) on the TS side
	// without a shared port table.
	nodeRow int32
	// layoutLinkTos holds the dst node ids of every LAYOUT double-link pair for which THIS
	// node is the SOURCE (alphabetically-first id — mirrors loader.go's emitLayoutLinks
	// de-dup rule, so each unordered pair streams from exactly one node's own fd, never
	// both). Sourced from LocalPolars (b.localPolars, the same LAYOUT model
	// computeLocalPolars/emitLayoutLinks use), set ONCE at construction (buildMoveDispatch)
	// since layout-link pairs are static after load — no per-cycle recompute. nil when
	// this node has no outbound layout-link pair (or in bare test construction).
	layoutLinkTos []string
	// nodeRowFor resolves a node id to its buffer NODE-ROW index (mirroring the old
	// central accumulator's NodeRowFor), injected via MoveDispatch.SetNodeStreams so this
	// package stays Buffer-independent. Used only to resolve this node's own
	// layoutLinkTos dst rows.
	nodeRowFor func(id string) (int32, bool)
	// edgeRowForPair resolves the buffer EDGE-ROW index of the bead edge connecting two
	// node ids (mirroring the old central accumulator's EdgeRowForPair), injected the same
	// way as nodeRowFor. -1/false when no bead edge connects the pair (the node-centers
	// fallback the overlay already handles for the combined block).
	edgeRowForPair func(a, b string) (int32, bool)
	// --- own selection/hover/abc-drag UI state (per-owner, no shared/republished map) ---
	//
	// This node's OWN current selected/hovered/latchedSel/gotDragMsg/dragDelta* bits —
	// set only by THIS node's own goroutine, from messages the gesture goroutine sends
	// on extIn (moveMsgKindSelect/Hover/Latched/AbcReset) or, for gotDragMsg/dragDelta*,
	// from this node's own neighborSetC handler when IT is the recipient of a peer's
	// drag (quantized_move.go's neighborSetCRequantize, dispatched via handle's
	// moveMsgKindNeighborSetC case — that already runs on this same goroutine). No
	// lock: only nm.handle (this goroutine) ever writes these, and writeStreamFrame
	// (also this goroutine) is the only reader.
	selected, hovered, latchedSel, gotDragMsg uint8
	dragDeltaA, dragDeltaB, dragDeltaC        int32
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
	buildFrame func(tick uint32, nodeRow int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, selected, kindID, hovered, latchedSel, gotDragMsg uint8, dragDeltaA, dragDeltaB, dragDeltaC int32, label string, portNames []string, portDX, portDY, portDZ, portPX, portPY, portPZ []float32, portIsInput, portHovered []uint8, dstNodeRows, edgeRows []int32, events []RowEvent) []byte
}

func newNodeMover(id string, geom nodeGeom, tr *T.Trace, clockSrc Clock) *nodeMover {
	// clk defaults to a fresh RealClock (its own independent origin — fine here: this
	// default is only ever read by a test that never launches run() as a goroutine;
	// production always overwrites it below with clockSrc.Copy() before the goroutine
	// does anything else), matching newEdgeMover's same default for the same reason.
	nm := &nodeMover{
		id: id, geom: geom,
		extIn: make(chan moveMsg, 8), neighborIn: map[string]chan moveMsg{}, tr: tr,
		partnerCenters: map[string]vec3{},
		clockSrc:       clockSrc, clk: NewRealClock(),
	}
	// Seed the atomic snapshot from the initial geometry (even when !HasPos, in which case
	// nodeWorldPos falls back to the origin) so readers — including another node's aimed-port
	// partnerCenter lookup — always have a valid center to read before the first center
	// message arrives.
	nm.snap.Store(&centerSnap{c: nodeWorldPos(geom), p: geom.ScenePolar, reach: geom.ReachR})
	return nm
}

// handle applies one move to this node: update held position, re-emit node-geometry.
func (m *nodeMover) handle(msg moveMsg) {
	if msg.NodeID != m.id {
		return
	}
	if msg.Kind == moveMsgKindAnchor {
		// Per-port anchor update: snap to ring-anchor index, mutate this node's held
		// port AnchorId, and re-emit node-geometry so the renderer redraws the port.
		ok := setPortAnchorId(&m.geom, msg.Port, msg.IsInput, msg.AnchorId)
		if !ok {
			return
		}
		if m.tr != nil {
			m.emitGeometry()
		}
		return
	}
	if msg.Kind == moveMsgKindCenter {
		// nodeMover is the SOLE writer of its own position (single-writer by
		// construction — this is the only path that mutates it). A Center payload is
		// the flat absolute-scene-polar drag write from fanCenters: apply it via
		// applyCenter, which also re-emits. A nil Center is fanCenters' PARTNER
		// re-emit (an aimed-port neighbor whose OWN center is unchanged, only asked
		// to re-emit so its port direction picks up the moved partner's fresh center
		// via m.partnerCenter at emit time) — no mutation, just re-emit.
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
		// always a FREE move now -- there is no equal-radii solve and no self-trigger
		// cascade to run.
		newPos := msg.Target
		if m.commitLocal != nil {
			m.commitLocal(m.id, newPos)
		}
		if m.tr != nil {
			m.tr.Breadcrumb("cascade.root", m.id, "", fmt.Sprintf("newPos=(%.4f,%.4f,%.4f)", newPos.X, newPos.Y, newPos.Z))
			// Structured buffer counterpart, riding this node's own dedicated
			// stream frame (emitGeometry's own next emit already fires from
			// commitLocal above, so this rides as a distinct events-only-shaped
			// write here rather than waiting on that one).
			m.writeStreamFrame([]RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbCascadeRoot, Debug: 1,
				NodeRow: m.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				X: newPos.X, Y: newPos.Y, Z: newPos.Z,
			}})
		}
		return
	}
	if msg.Kind == moveMsgKindDragStart {
		m.armDragAnchor()
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
	if msg.Kind == moveMsgKindAbcReset {
		m.gotDragMsg = 0
		m.dragDeltaA, m.dragDeltaB, m.dragDeltaC = 0, 0, 0
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
			m.emitGeometry()
		}
		return
	}
	if msg.Kind == moveMsgKindNeighborSetC {
		// Neighbor edge re-quantize (receiver-computes, one hop, no forward): SenderID
		// (the dragged node) moved to msg.FromCenter; THIS node stays put and re-quantizes
		// its OWN edge to SenderID from the live offset — theta, phi AND r all fresh —
		// so both the angle and the distance to SenderID change (neighborSetCRequantize).
		if m.neighborSetC != nil {
			m.neighborSetC(m.id, msg.SenderID, nodeWorldPos(m.geom), msg.FromCenter, msg.DeltaA, msg.DeltaB, msg.DeltaC)
		}
		return
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

// armDragAnchor (re-)arms this node's drag-anchor snapshot from its CURRENT
// LocalPolar triples — always overwriting whatever was there, so a new drag on this
// same node re-arms rather than keeping a stale anchor from the previous drag. Runs
// only on this node's own goroutine (moveMsgKindDragStart handler). See
// moveMsgKindDragStart's doc comment for why this fires exactly once per drag.
func (m *nodeMover) armDragAnchor() {
	byTo := map[string]LocalPolar{}
	if m.layoutHolderFn != nil {
		if lh := m.layoutHolderFn(); lh != nil {
			for _, lp := range lh.LocalPolarsSnapshot() {
				byTo[lp.To] = lp
			}
		}
	}
	m.dragAnchorByTo = byTo
	m.dragAnchorArmed = true
}

// applyCenter is the SOLE WRITE of this node's center/reach. It is called ONLY from
// this nodeMover's own inbox-drain goroutine (handle's moveMsgKindCenter case, driven
// by fanCenters below), which is what makes that one goroutine the exclusive writer of
// m.geom/m.snap. It sets the held polar position, publishes the atomic snapshot readers
// observe cross-goroutine (stdin reader: centerOfNode/heldCenters/heldPolar/fanCenters'
// partner lookup, edgeMover's partnerCenter), and re-emits this node's live geometry.
func (m *nodeMover) applyCenter(center vec3, reach float64) {
	setNodeWorld(&m.geom, center)
	m.geom.ReachR = reach
	m.snap.Store(&centerSnap{c: center, p: m.geom.ScenePolar, reach: reach})
	// Push this fresh center to every direct neighbor (nm.neighborIn's key set — one
	// hop, no cascade) so each neighbor's OWN partnerCenters map picks it up via
	// moveMsgKindNeighborCenter (handle, below) — the delivery-mechanism replacement
	// for the old cross-goroutine atomic snap read. Routed through m.sendMove (this
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

// emitGeometry re-emits this node's authoritative geometry. A CONNECTED port marker is
// AIMED at its partner's current center (m.partnerCenter, atomic-snapshot-backed); an
// edgeless port falls back to its own polar-torus ring-anchor placement (portWorldPos).
// This method, applyCenter, and setPortAnchorId (via handle) all run on
// nodeMover's own inbox-drain goroutine only (see the doc comment on nodeMover.geom),
// so a plain field read here can never race a concurrent writer.
func (m *nodeMover) emitGeometry() {
	// Dedicated per-node stream (see streamOut's doc comment): write this node's own
	// combined frame immediately on a geometry change, in addition to the tick-driven
	// write in run()'s loop (mirrors edgeMover.recomputeGeometry's writeStreamFrame call).
	// NodeGeometry rides THIS frame's own EVENTS section (fully decentralized — it
	// never rides the VIEW stream's fallback bucket) — this
	// nodeMover is the sole owner of its node's geometry, so it resolves its own
	// NodeRow at the call site (owner_events.go) rather than routing through a
	// shared accumulator.
	m.writeStreamFrame([]RowEvent{{
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
func (m *nodeMover) writeStreamFrame(events []RowEvent) {
	if m.streamOut == nil || m.buildFrame == nil {
		return
	}
	center := nodeWorldPos(m.geom)
	sphereR := effectiveRadius(m.geom)
	label := m.geom.Label
	if label == "" {
		label = m.id
	}
	portPosDir := aimedPortPosDir(m.geom, m.partnerCenter)
	ports := buildPortGeoms(m.geom, portPosDir)
	portNames := make([]string, len(ports))
	portDX := make([]float32, len(ports))
	portDY := make([]float32, len(ports))
	portDZ := make([]float32, len(ports))
	portPX := make([]float32, len(ports))
	portPY := make([]float32, len(ports))
	portPZ := make([]float32, len(ports))
	portIsInput := make([]uint8, len(ports))
	portHovered := make([]uint8, len(ports))
	for i, p := range ports {
		portNames[i] = p.Name
		portDX[i], portDY[i], portDZ[i] = float32(p.DX), float32(p.DY), float32(p.DZ)
		portPX[i], portPY[i], portPZ[i] = float32(p.PX), float32(p.PY), float32(p.PZ)
		if p.IsInput {
			portIsInput[i] = 1
		}
		if m.hovered == 1 && m.hoverPort != "" && m.hoverPort == p.Name && m.hoverIsInput == p.IsInput {
			portHovered[i] = 1
		}
	}
	selected, hovered, latchedSel, gotDragMsg, kindID := m.selected, m.hovered, m.latchedSel, m.gotDragMsg, m.kindID
	dA, dB, dC := m.dragDeltaA, m.dragDeltaB, m.dragDeltaC
	// This node's own outbound layout-links (layoutLinkTos, static since load — see its
	// doc comment): resolve each dst id to its CURRENT buffer node row + the CURRENT bead
	// edge row connecting the pair (both re-resolved every emit, mirroring the combined
	// block's edgeRowForPair re-resolve every emit). A dst id that hasn't
	// registered a node row yet is skipped (mirrors resolvableLayoutLinks' endpoint
	// filter) rather than packed with a -1 dst row.
	var dstNodeRows, edgeRows []int32
	if len(m.layoutLinkTos) > 0 && m.nodeRowFor != nil {
		dstNodeRows = make([]int32, 0, len(m.layoutLinkTos))
		edgeRows = make([]int32, 0, len(m.layoutLinkTos))
		for _, to := range m.layoutLinkTos {
			dstRow, ok := m.nodeRowFor(to)
			if !ok {
				continue
			}
			edgeRow := int32(-1)
			if m.edgeRowForPair != nil {
				if r, ok := m.edgeRowForPair(m.id, to); ok {
					edgeRow = r
				}
			}
			dstNodeRows = append(dstNodeRows, dstRow)
			edgeRows = append(edgeRows, edgeRow)
		}
	}
	frame := m.buildFrame(uint32(m.clk.Tick()), m.nodeRow,
		float32(center.X), float32(center.Y), float32(center.Z),
		float32(nodeRadius(m.geom.Kind)), float32(sphereR),
		verticalRingNormalX, verticalRingNormalY, verticalRingNormalZ,
		flatRingNormalX, flatRingNormalY, flatRingNormalZ,
		selected, kindID, hovered, latchedSel, gotDragMsg, dA, dB, dC,
		label, portNames, portDX, portDY, portDZ, portPX, portPY, portPZ, portIsInput, portHovered,
		dstNodeRows, edgeRows, events)
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
		ApplySpeedNonBlocking(m.clk, m.speedCh)
		// Drain every dedicated inbound channel non-blockingly, repeating until a
		// full pass yields nothing — this is the "drain to empty, don't throttle a
		// backlog" half of the shape.
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
