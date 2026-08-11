// node_geometry_accessors.go — the POST-CONSTRUCTION exported surface package Wiring
// reaches repeatedly while a node's own goroutine is running (as opposed to
// node_geometry_wire.go's construction-time-only wiring methods).
// docs/planning/movedispatch-decomposition.md §20.
//
// The channel-touching members (msg.centerOut, msg.extIn, msg.neighborIn) stay unexported
// and are reached ONLY through the methods below that close over them — PollCenter,
// SendExternal, EnqueueSend, NeighborTrySend — mirroring §17's edgeMover rule exactly: no
// channel is ever exported, so no other package can ever send on this node's own inbox by
// naming the channel directly.
package nodeactor

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// ID returns this node's own id.
func (m *NodeGeometry) ID() string { return m.id }

// Traced reports whether this node has a live *Trace.Trace wired (false in a bare/headless
// test build) — the guard package Wiring's commit path uses before building an expensive
// diagnostic breadcrumb string, mirroring every other `if m.tr != nil` gate this package's
// own files use internally.
func (m *NodeGeometry) Traced() bool { return m.tr != nil }

// Breadcrumb emits a DEBUG breadcrumb through this node's own *Trace.Trace, a no-op when
// Traced() is false. Wraps m.tr.Breadcrumb so package Wiring's commit path
// (commitNodeMoveLocal's "bead-crud" diagnostic) never reaches the unexported tr field
// directly — the same channel/handle-wrapping treatment this file gives every other
// cross-goroutine touch.
func (m *NodeGeometry) Breadcrumb(label, node, port, value string) {
	if m.tr != nil {
		m.tr.Breadcrumb(label, node, port, value)
	}
}

// Kind returns this node's own kind name — a field on the embedded, write-once
// NodeIdentity (see this type's own doc comment on why that split makes this call safe
// from any goroutine).
func (m *NodeGeometry) Kind() string { return m.geom.Kind }

// SelfKind returns this node's own kind name as SEEDED BY THE SPEC (SetSelfKind,
// node_geometry_wire.go) — read by package Wiring's beadcrud adapter
// (dragTouchingBeads/touching_beads.go). Distinct field from Kind() above (geom.Kind is
// the load-time NodeIdentity's own copy); both carry the same value in production, kept as
// two fields because they were seeded on two different paths before this move and neither
// call site was worth re-plumbing to share one.
func (m *NodeGeometry) SelfKind() string { return m.selfKind }

// Tick returns this node's own clock's current tick — read-only, used by package Wiring's
// clock-speed regression coverage (pair_node_self_clock_speed_test.go) to measure how far
// this node's own render clock advances over a wall-clock window, the same "measure tick
// advance, not the raw Clock" shape clock_speed_test.go already uses.
func (m *NodeGeometry) Tick() int64 { return m.clocks.clk.Tick() }

// Label returns this node's own load-time label (empty when the spec set no data.label —
// callers fall back to the id themselves, the same convention writeStreamFrame's own
// label-or-id fallback uses).
func (m *NodeGeometry) Label() string { return m.geom.Label }

// WorldCenter returns this node's own current world-space center
// (sceneCenter + polar2cart(scenePolar)) — same computation node_geometry_stream.go's own
// writeStreamFrame/chainBeads make, read-only, safe from this node's own goroutine only
// (m.geom's mutable half is single-writer — see this type's own doc comment).
func (m *NodeGeometry) WorldCenter() vec3 { return nodegeom.NodeWorldPos(m.geom) }

// NodeRow returns this node's own stable buffer row (WireStream's row argument,
// node_geometry_wire.go) — used by package Wiring's commit path to stamp a foreign-row
// breadcrumb's own NodeRow column (see WriteStreamFrame's doc comment on the
// ownership-vs-reference split).
func (m *NodeGeometry) NodeRow() int32 { return m.stream.nodeRow }

// EdgeIDs returns this node's own incident-edge id list (AddEdgeID, node_geometry_wire.go)
// — read-only; callers only ever range it (package Wiring's commitNodeMoveLocal,
// dragTouchingBeads).
func (m *NodeGeometry) EdgeIDs() []string { return m.topo.edgeIDs }

// PartnerCenters returns this node's own live copy of every direct neighbor's last-known
// world center (map keyed by neighbor id) — read-only; callers only ever look a key up or
// range it (package Wiring's commitNodeMoveLocal, dragTouchingBeads). Never nil after
// NewNodeGeometry.
func (m *NodeGeometry) PartnerCenters() map[string]vec3 { return m.topo.partnerCenters }

// NeighborKinds returns this node's own direct-neighbor-id → kind-name map (AddNeighborKind,
// node_geometry_wire.go) — read-only, handed wholesale to beadcrud.DragTouchingBeads by
// package Wiring's dragTouchingBeads adapter.
func (m *NodeGeometry) NeighborKinds() map[string]string { return m.topo.neighborKinds }

// SendMove returns this node's own bound outbound-retry-queue send func (WireMessaging's
// sendMove, node_geometry_wire.go) — package Wiring's commitNodeMoveLocal hands it straight
// into broadcastToEdgesAndPartners as the enqueue func every fan-out message goes through.
// A func value, not a channel: the same bound-func-value handoff shape this whole file
// uses throughout.
func (m *NodeGeometry) SendMove() func(id string, msg movemsg.Msg) { return m.msg.sendMove }

// NeighborIDs returns this node's own direct-neighbor id set (the key set of its
// neighborIn channel map, EnsureNeighborChannel) as a plain slice — construction-time only
// (package Wiring's move_dispatch_construct.go partnerCenters seed loop, single-threaded,
// before any driving goroutine exists).
func (m *NodeGeometry) NeighborIDs() []string {
	ids := make([]string, 0, len(m.msg.neighborIn))
	for id := range m.msg.neighborIn {
		ids = append(ids, id)
	}
	return ids
}

// NeighborTrySend returns the non-blocking try-send func FROM peer fromID TO this node —
// wrapping this node's own neighborIn[fromID] channel exactly like edgemover's
// TrySendFromSrc/TrySendFromDst wrap that actor's own channels, so the channel itself never
// crosses the package boundary. Called by package Wiring's move_dispatch_construct.go
// resolveDest closure (bound at construction, invoked later — every invocation runs on
// fromID's own driving goroutine, via that node's own flushPending). ok is false when this
// node has no dedicated channel for fromID (not a direct neighbor).
func (m *NodeGeometry) NeighborTrySend(fromID string) (func(movemsg.Msg) bool, bool) {
	ch, ok := m.msg.neighborIn[fromID]
	if !ok {
		return nil, false
	}
	return func(msg movemsg.Msg) bool {
		select {
		case ch <- msg:
			return true
		default:
			return false
		}
	}, true
}

// PollCenter drains this node's own centerOut channel non-blockingly, returning the latest
// pushed center (ApplyCenter's own latest-wins push) if one is waiting. Called ONLY from
// package Wiring's dispatch/gesture goroutine (moverRegistry.drainCenterMirror) — the sole
// reader of every node's centerOut channel; this node's own goroutine is the sole writer
// (ApplyCenter).
func (m *NodeGeometry) PollCenter() (vec3, bool) {
	select {
	case c := <-m.msg.centerOut:
		return c, true
	default:
		return vec3{}, false
	}
}

// SendExternal sends msg on this node's own dedicated external-entry channel (extIn) — the
// bare EXTERNAL-caller path (package Wiring's RootMove/drag, gesture.go's dragStart send),
// not a mover-to-mover send (those go through EnqueueSend onto a node's own pending/
// flushPending retry queue, never through this method). Blocking with a ctx-cancel escape
// hatch: without the ctx.Done() arm, a send into a torn-down/full extIn on shutdown parks
// the calling goroutine forever (this node's own Run loop has already returned on the same
// ctx cancel, so nothing will ever drain it). ctx is nil only in tests that build a bare
// MoveDispatch without Start — a nil Context's Done() channel would panic, so guard it and
// fall back to a plain blocking send there (matches prior behavior; no shutdown path
// exists in that setting anyway). This bare path never fires the test-only tap — see
// EnqueueSend's own doc comment for the one that does.
func (m *NodeGeometry) SendExternal(ctx context.Context, msg movemsg.Msg) {
	if ctx == nil {
		m.msg.extIn <- msg
		return
	}
	select {
	case m.msg.extIn <- msg:
	case <-ctx.Done():
	}
}

// TryRecvExternal returns the next message on this node's own external-entry inbox
// (extIn), non-blocking — the receive-side counterpart to SendExternal, needed only when
// no NodeMover.Run/PairNodeSelf.Step goroutine is draining this node's own inbox (a bare
// test construction with no driving goroutine started). Production drains extIn from
// inside Run/Step's own select loop, never through this method.
func (m *NodeGeometry) TryRecvExternal() (movemsg.Msg, bool) {
	select {
	case msg := <-m.msg.extIn:
		return msg, true
	default:
		return movemsg.Msg{}, false
	}
}

// EnqueueSend is THIS node's own non-blocking send: it fires this node's own tap (at
// enqueue time, so tap-based tests' counts/ordering match today's behavior — a plain nil
// check + direct call, since m.msg.tap is owned and read only by this node's own
// goroutine, the only caller of this method once bound — see SendMove's own doc comment),
// appends the message to this node's own pending retry queue, and attempts an immediate
// flush — never blocking the calling handler goroutine. Bound once per node, at
// construction, as this node's own m.msg.sendMove (package Wiring's
// move_dispatch_construct.go: `ng.WireMessaging(..., md.mr.enqueueFor(ng), ...)`, where
// enqueueFor now just returns this method value) so every send this node's own handle
// performs — including the ones broadcastToEdgesAndPartners makes on this node's own
// behalf via SendMove() — goes through this node's own retry queue, never a raw blocking
// channel write and never a second node's queue.
//
// This absorbs what used to be package Wiring's mover_registry.go enqueueFor closure body
// (a direct external touch of msg.tap/msg.pending) into the actor itself
// (docs/planning/movedispatch-decomposition.md §20) — the same class of runtime,
// own-goroutine write §19 already left as a bare field write for quantOffset/pending, now
// made unreachable from outside this package by construction rather than by convention.
func (m *NodeGeometry) EnqueueSend(destID string, msg movemsg.Msg) {
	if m.msg.tap != nil {
		m.msg.tap(destID, msg)
	}
	m.msg.pending = append(m.msg.pending, pendingSend{destID: destID, msg: msg})
	m.flushPending()
	if len(m.msg.pending) > maxPendingSends {
		// Named causes, checked against flushPending's actual behaviour (not
		// guessed): an item whose destID doesn't resolve is DROPPED, not
		// retained (flushPending's `!ok` branch), so an unresolvable
		// destination can never grow this queue — it is deliberately not
		// named below. What CAN: (1) a peer whose own goroutine has
		// stopped draining its inbox entirely (wedged or dead) — every
		// later item to that same destination piles up behind it,
		// unattempted, to preserve FIFO; (2) this node enqueueing to a
		// live-but-slower peer faster, cycle over cycle, than that peer's
		// own goroutine drains its inbox — flushPending retries only ONE
		// send per blocked destination per cycle, so a persistent
		// per-cycle surplus accumulates even without a dead peer.
		panic(fmt.Sprintf(
			"NodeGeometry(%s): pending exceeded %d retry-queued sends; either a "+
				"destination's own goroutine has stopped draining its inbox "+
				"(wedged or dead), or this node is enqueueing to a peer faster "+
				"than that peer drains, cycle over cycle",
			m.id, maxPendingSends))
	}
}

// QuantOffset returns this node's own current quantized polar offset triple
// (iTheta,iPhi,iR) — read by package Wiring's moverRegistry.nodeQuantOffset (external
// verification, e.g. confirming a real reload lands on the same offset a live edit just
// persisted).
func (m *NodeGeometry) QuantOffset() (iTheta, iPhi, iR int) {
	return m.quantOffset.ITheta, m.quantOffset.IPhi, m.quantOffset.IR
}

// QuantizedOffsetValue returns this node's own current quantized polar offset as the full
// quantoffset.QuantizedOffset (including its per-node step constants, cTheta/cPhi/cR),
// for callers that need to feed it back into quantoffset.DeriveCenters or similar —
// unlike QuantOffset's plain int triple, this preserves a non-default per-node step if one
// were ever set (none is, today).
func (m *NodeGeometry) QuantizedOffsetValue() quantoffset.QuantizedOffset { return m.quantOffset }

// ReachR returns this node's own current reach radius (nodegeom.NodeGeom.ReachR — the max
// distance to any node it outputs to, computed at load by topoderive.ReachRFromPolar and
// re-derived on every commit). Read-only external verification (package Wiring's
// build_load_derive_test.go).
func (m *NodeGeometry) ReachR() float64 { return m.geom.ReachR }

// CommitQuantOffset measures this node's fresh quantized offset against committedPolar,
// stores it (the RUNTIME write — this node's own driving goroutine, on every drag commit;
// distinct from SetQuantOffset's construction-time seed, node_geometry_wire.go), and
// persists it. This is package Wiring's commitNodeMoveLocal's own 3-statement block
// (measure/store/persist), folded into one method in §20
// (docs/planning/movedispatch-decomposition.md) so the runtime write to quantOffset never
// needs an exported setter — the same treatment §19 gave the sibling msg.pending write via
// EnqueueSend above.
func (m *NodeGeometry) CommitQuantOffset(committedPolar geom.Polar) {
	off := quantoffset.MeasureScalar(committedPolar, m.quantOffset)
	m.quantOffset = off
	m.persistQuantOffset(off, committedPolar)
}

// WriteStreamFrame writes an out-of-band frame (typically one carrying only breadcrumb
// events) onto this node's own dedicated stream — the exported door to writeStreamFrame
// (node_geometry_stream.go) package Wiring's commit path uses to ride a breadcrumb on this
// node's own frame (commitNodeMoveLocal's "bead-crud" diagnostic). Must be called only
// from this node's own driving goroutine, exactly like every other writeStreamFrame call.
func (m *NodeGeometry) WriteStreamFrame(events []wire.RowEvent) {
	m.writeStreamFrame(events)
}
