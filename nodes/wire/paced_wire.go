package wire

import (
	"context"
	"fmt"
	"math"
	"os"

	T "github.com/dtauraso/wirefold/Trace"
)

// edgeBeadTraceEnabled gates whether stepAll appends a T.KindEdgeBead pendingWireEvent
// for every in-flight bead every tick. It is read ONCE at process startup from the
// WIREFOLD_EDGE_BEAD_TRACE env var (same "one env var, read once before any wire
// goroutine exists" shape as Buffer/stream_fds.go's WIREFOLD_STREAM_FDS) -- never
// re-read per tick, so this package-level bool is race-free by construction (written
// once at init, before any PacedWire goroutine starts; see
// memory/feedback_no_atomics_are_defects.md). Default (env var absent/unset/anything
// but "1") is OFF: KindEdgeBead is a high-volume per-tick-per-bead event whose sole
// consumer is the opt-in .probe trace log (wirefold.probe.trace, default false) --
// LiveBeadRow / the Bead-block buffer path that actually RENDERS beads does not read
// this flag and is unaffected. KindBreadcrumb and KindArrive are NOT gated by this
// flag and always emit.
var edgeBeadTraceEnabled = os.Getenv("WIREFOLD_EDGE_BEAD_TRACE") == "1"

// wireChanBufferSize bounds PacedWire's in-channel (source -> wire) and out-channel
// (wire -> destination). Generously sized so the SOURCE's send (Send) and the WIRE's
// own delivery send (inside driveOneCycle) can never need to block (MODEL.md
// "Sending": "no back-pressure, ever" -- the send must always succeed immediately).
// A node fires at most a handful of times between two wire-drive cycles (~16ms
// apart), so this ceiling is never approached under any realistic load; if it ever
// were, that is a bug to fix at the source (firing far faster than any wire could
// ever drain), not a reason to make either send blocking.
const wireChanBufferSize = 4096

// deliveredBead is a value the wire's own goroutine has finished timing and handed
// off toward the destination over outCh. deliverTick is the tick THIS WIRE's own
// clock reading was at when the handoff happened -- the authoritative "what tick did
// this actually land on" answer (RecvTick/PollRecvTick's tick return).
type deliveredBead struct {
	val         int
	deliverTick int64
}

// PulseSpeedWuPerMs is the fixed world-units-per-MILLISECOND conversion for the
// SimLatencyMs REPORTING path (the ms value emitted on the send trace); it is NOT
// the clock's unit. This is an intentional duplicate of the literal value in
// nodes/Wiring/curve_params.go's CurveParamPulseSpeedWuPerMs — that copy is the
// single source of truth gen-node-defs reads (by literal AST value) to emit TS,
// and it cannot be an alias of this one because wire must not import Wiring
// (that would be a package cycle: Wiring already imports wire). Keep the two
// literals in sync by hand; a mismatch would only affect SimLatencyMs
// reporting/pacing math, not correctness of delivery.
const PulseSpeedWuPerMs = 0.04

// PulseSpeedWuPerTick is the uniform pulse speed reinterpreted in world-units per
// TICK (MODEL.md: pulseSpeed is world-units-per-tick). It is the ms speed scaled
// by the tick period: 0.04 wu/ms × 16 ms/tick = 0.64 wu/tick. This is what the
// human-speed clock uses to derive ticksToCross = arcLength / PulseSpeedWuPerTick,
// which equals the retired arc/pulseSpeedMs/16 sample count — so a bead visits the
// same number of positions in the same wall time.
const PulseSpeedWuPerTick = PulseSpeedWuPerMs * MsPerTick

// beadPlacement bundles everything one placement needs. Steps times delivery
// (docs/bead-lattice.md "Timing" — ticksToCross = steps * dwell, no length to
// divide); the segment endpoints + source identity drive the per-frame position
// stream. Geometry travels WITH the bead, never stored on the shared wire, so
// each in-flight bead evaluates the exact segment it is drawn on.
// The zero value (empty segment + identity) means "no position stream" — unit
// tests that only exercise delivery pass just Steps.
type beadPlacement struct {
	// Steps is this placement's bead-step count (docs/bead-lattice.md "The
	// count") — copied straight from the sending Out's own Geom().Steps at Send
	// time, so the wire never re-derives it from anything else.
	Steps int
	// Position-stream context. Start/End are this edge's straight-segment endpoints
	// (source OUT-port world pos, dest IN-port world pos). Node/Port are the SOURCE
	// node id + output port — the position trace key, matching the send event so the
	// renderer routes by source+sourceHandle (fan-out).
	Start, End Vec3
	Node, Port string
}

// streams reports whether this placement carries position-stream context. False
// for the bare-delivery placements used by unit tests (empty Node).
func (bp beadPlacement) streams() bool {
	return bp.Node != ""
}

// placeRequest is what Send hands across the wire's in-channel: a bead value,
// the placement geometry it should be timed/drawn against, and the tick to
// stamp it with. The SENDING node's own goroutine reads its own clock ONCE per
// emission and stamps placementTick here (MODEL.md's Clock bullet: "placement
// is decided by the emitting goroutine, at the moment it calls Send, from its
// own clock, once per emission — not re-derived later by whichever goroutine
// happens to drain the wire's in-channel"). The wire itself is a passive
// struct with no goroutine of its own — it is drained by its source node's
// mover (node_mover.go) — so it is no longer the reader. Moving the read to
// the sender is what lets several beads placed in the same emission (e.g. a
// broadcast fan-out) provably share one placementTick: reading a fresh clock
// value per wire in the drain pass could straddle a tick boundary, splitting
// one emission across two ticks.
type placeRequest struct {
	val           int
	bp            beadPlacement
	placementTick int64
}

// inflightBead is one bead traversing the wire. Each bead carries its own
// geometry so a mid-flight geometry edit (node-move) re-derives the remaining
// travel from the NEW step count while preserving the bead's FRACTIONAL progress t.
// Distance is NOT stored: fractional progress t = (clock.Tick() − placementTick)
// / ticksToCross is a pure function of the wire's own single clock reading
// (MODEL.md "Geometry and time").
//
// Every field here is touched by EXACTLY ONE goroutine: this wire's own (folded
// into edgeMover.run via driveOneCycle — see the PacedWire doc comment below).
type inflightBead struct {
	val           int
	placementTick float64     // this wire's own tick reading when placed (fractional after a geometry rebase)
	steps         int         // current bead-step length of this bead's edge (docs/bead-lattice.md)
	seg           WireSegment // current straight-segment endpoints of this bead's edge
	node          string      // source node id — the position/cancel routing key
	port          string      // source output port — the position/cancel routing key
	streams       bool        // whether this bead carries position-stream context
	gen           uint64      // per-bead id (the bead's emitted identity)
	// finalPending is true once a drive cycle has advanced this bead to its
	// delivery deadline (target==deadline) but the handoff to outCh has not yet
	// succeeded (e.g. it was not yet at the FIFO head, or outCh had no room that
	// cycle). Subsequent cycles retry only the (cheap) delivery handoff for such
	// a bead, without re-running the position-advance math.
	finalPending bool
}

// ticksToCross returns the tick count for a bead of the given STEP count to
// cross, at this wire's own dwell-per-bead: steps * dwell (docs/bead-lattice.md
// "Timing" — a longer edge is simply more beads, dwell is a constant, so there
// is no per-edge division left the way arcLength/pulseSpeed used to require).
// Fractional; the driver delivers on the first integer tick at or past
// placementTick + this.
func (pw *PacedWire) ticksToCross(steps int) float64 {
	if pw.dwell <= 0 {
		return 0
	}
	return float64(steps) * pw.dwell
}

// NextArrivalTick is the tick the EARLIEST in-flight bead on this wire will finish
// crossing on — known at PLACEMENT, not discovered by watching: a bead placed at
// placementTick with a step count of steps lands at placementTick + steps*dwell, and
// nothing about that changes while it flies. ok is false when nothing is in flight.
//
// This is what lets the source node SLEEP to the moment a traversal completes instead of
// waking every cycle to ask whether it has. The animation is unaffected either way —
// LiveBeadFractions is a pure function of (current tick, placementTick), so the pulse's
// drawn position never depended on this goroutine being awake.
//
// Called on this wire's own goroutine, like every other reader of inflight.
func (pw *PacedWire) NextArrivalTick() (int64, bool) {
	best := int64(0)
	found := false
	for _, b := range pw.inflight {
		// Ceil: delivery happens on the first INTEGER tick at or past the deadline
		// (ticksToCross' own doc comment), so the wake must not land a fraction early.
		at := int64(math.Ceil(b.placementTick + pw.ticksToCross(b.steps)))
		if !found || at < best {
			best, found = at, true
		}
	}
	return best, found
}

// EarliestArrival is the SOONEST arrival across n wires — what a node with several
// outgoing edges sleeps to. The shorter edge wins because it is the first thing that needs
// this goroutine awake; a longer one simply has not arrived yet, and waking for it early is
// the polling this replaces. ok is false when nothing is in flight on any of them.
//
// A free function over a slice rather than a method on the node: the node owns which wires
// it drives, and this owns none of them — it only reads each one's own answer.
func EarliestArrival(wires []*PacedWire) (int64, bool) {
	best := int64(0)
	found := false
	for _, w := range wires {
		if w == nil {
			continue
		}
		if at, ok := w.NextArrivalTick(); ok && (!found || at < best) {
			best, found = at, true
		}
	}
	return best, found
}

// PacedWire is an ACTIVE GOROUTINE (MODEL.md "The network"), not a passive
// struct: a channel in from its source node, a channel out to its destination
// node. It is NOT a separate goroutine of its own — it is driven by the same
// per-edge goroutine that already existed to revise in-flight geometry on a
// node-move (edgeMover.run, node_mover.go), which owns the beads it revises.
//
//   - inCh is the wire's IN-CHANNEL: the source node's own goroutine calls Send,
//     a non-blocking buffered-channel send, and moves on (MODEL.md "Sending" — no
//     back-pressure, ever).
//   - outCh is the wire's OUT-CHANNEL: the destination node's own goroutine calls
//     RecvTick/Recv, a non-blocking buffered-channel receive.
//   - inflight/nextGen/dwell are owned EXCLUSIVELY by the wire's
//     own goroutine (driveOneCycle, called every cycle from edgeMover.run) — exactly
//     one writer and one reader (the same goroutine).
type PacedWire struct {
	inCh  chan placeRequest
	outCh chan deliveredBead

	// Owned exclusively by this wire's own goroutine (driveOneCycle and its
	// helpers, and ReviseInFlightGeometry — both called only from edgeMover.run,
	// which IS this wire's goroutine).
	inflight []inflightBead
	// nextGen mints a unique id for each placed bead (the bead's emitted identity).
	nextGen uint64
	// dwell is this wire's own ticks-per-bead-step (docs/bead-lattice.md
	// "Timing"): ticksToCross(steps) = steps * dwell. Same TEST-affordance role
	// the retired pulseSpeed field had (see NewPacedWire's doc comment) — the
	// one production call site passes DwellTicksPerBead.
	dwell float64

	Target       string // destination node id — the wire's destination routing key
	TargetHandle string // destination input-port name — the wire's destination routing key

	// readout is this wire's REPORTING apparatus — the pending Position/Arrive
	// buffer, the Trace handle, the debug-breadcrumb channel and its drop counter
	// (wire_readout.go). It is a separate type because none of it is transport: the
	// queue above decides when a bead arrives, the readout only says so. Ownership is
	// unchanged by that separation — the same single goroutine owns each of its
	// fields as owned them when they sat flat here (see wireReadout's doc comment).
	readout wireReadout
}

// maxInflightBeads is the declared upper bound on len(pw.inflight) between
// FIFO-head deliveries (docs/planning/visual-editor/session-log.md Step 3, "inflight"). stepAll delivers
// ONLY from the FIFO head (i==0) onto outCh; if the destination has stopped
// draining outCh, the head cannot hand off and nothing behind it can either
// (stepAll's own doc comment), so inflight can only grow while the stall
// persists. In this model every node goroutine always drains its inputs
// (MODEL.md), so a destination that has permanently stopped draining is a
// BUG — a stalled or dead destination goroutine — never legitimate
// backpressure; there is no "the destination is just slow" case to
// distinguish from "the destination is gone".
//
// Like maxPendingEvents, this is a GENEROUS ceiling, not a tight derivation,
// and honestly so: placements can only enter inflight via drainPlacements
// draining inCh, and wireChanBufferSize already ceilings how many can ever
// be outstanding there, so reusing it here gives a value definitely wrong to
// reach rather than one derived airtight from a bounded steady-state depth.
const maxInflightBeads = wireChanBufferSize

// NewPacedWire creates an empty PacedWire with its in/out channels ready. steps
// is UNUSED by this constructor (the bead-step count lives per-bead, on each
// placement, not on the wire itself) — it survives as a parameter only so a
// call site reads as self-documenting length-then-speed, the same shape
// arcLength/pulseSpeed had (arcLength was equally unused there). dwellTicks is
// this wire's own ticks-per-bead-step (use DwellTicksPerBead in production).
//
// PULSE SPEED IS UNIFORM ACROSS ALL WIRES — per-wire speed is rejected doctrine, and the
// TS layer cannot even express it (no speed prop in WireProps). The dwellTicks PARAMETER
// survives only as a TEST affordance: the lean per-node tests pass 1.0 with steps set to
// the desired latMs so that ticksToCross falls out as latMs. What keeps production uniform
// is that there is exactly ONE non-test call site (loader.go), passing DwellTicksPerBead.
//
// That one-call-site invariant is enforced by tools/check-uniform-pulse-speed.sh. Do not
// add a second production caller: it converts "uniform" from structural to conventional.
// If production ever needs to build a wire elsewhere, drop this parameter instead and let
// the tests express ticksToCross directly as steps*DwellTicksPerBead.
func NewPacedWire(steps int, dwellTicks float64) *PacedWire {
	return &PacedWire{
		dwell:   dwellTicks,
		inCh:    make(chan placeRequest, wireChanBufferSize),
		outCh:   make(chan deliveredBead, wireChanBufferSize),
		readout: wireReadout{breadcrumbCh: make(chan RowEvent, 4)},
	}
}

// SendOutcome distinguishes WHY Send did not place a bead, so a caller cannot
// accidentally treat "buffer momentarily full" (transient — never exit on this)
// the same as a genuinely terminal condition. There is currently only one
// non-terminal failure mode (inCh full); SendOutcome exists as a TYPE so that
// if a real terminal wire-teardown path is ever reintroduced, it lands as a
// third, distinct constant rather than silently widening SendBufferFull's
// meaning.
type SendOutcome uint8

const (
	// SendPlaced: the bead was enqueued onto inCh.
	SendPlaced SendOutcome = iota
	// SendBufferFull: inCh's buffer (wireChanBufferSize) was full at send time.
	// TRANSIENT, NOT TERMINAL — the wire's own goroutine drains inCh every
	// cycle, so room reappears almost immediately. A caller must NOT exit its
	// drive loop on this outcome; it should skip this cycle's placement and
	// retry on the next one. See the wireChanBufferSize doc comment: this
	// should never occur under realistic load, but if it does, the source
	// keeps running rather than silently losing its drive goroutine forever.
	SendBufferFull
)

// Send enqueues one bead placement onto this wire's IN-CHANNEL from the SOURCE
// node's own goroutine (Out.placeDrivenNoWalker). Non-blocking by construction:
// the buffered channel means this call always succeeds immediately under any
// realistic load, so the source never waits on the wire or the destination
// (MODEL.md "Sending" — no back-pressure, ever). Returns SendBufferFull only if
// the (generously sized) buffer is somehow already full — a condition that would
// itself indicate a bug elsewhere (a source firing far faster than any wire could
// ever drain), never ordinary traffic. Emits ONE breadcrumb per occurrence (this
// should never fire; it is a control-event signal, not a per-tick firehose) —
// see CLAUDE.md's debug-breadcrumb section.
//
// tick is the SENDER's own clock reading for this placement (read ONCE by the
// caller, not by the wire — see placeRequest's doc comment).
func (pw *PacedWire) Send(v int, bp beadPlacement, tick int64) SendOutcome {
	// Report any breadcrumb drops from a PREVIOUS call before doing anything
	// else this call — see flushDroppedBreadcrumbs' doc comment. A no-op when
	// nothing has been dropped.
	pw.readout.flushDroppedBreadcrumbs()
	select {
	case pw.inCh <- placeRequest{val: v, bp: bp, placementTick: tick}:
		return SendPlaced
	default:
		if pw.readout.Trace != nil {
			pw.readout.Trace.Breadcrumb("wire-send-buffer-full", pw.Target, pw.TargetHandle, "")
		}
		// Structured buffer counterpart (memory/feedback_no_single_writer_bridge.md):
		// non-blocking send of the dropped bead's value, drained by this wire's own
		// goroutine (edgeMover.writeStreamFrame's drainBreadcrumbEvents call) — never
		// blocks the source node, and drops the diagnostic (not the bead itself,
		// which was never dropped either — SendBufferFull is transient/retryable) if
		// breadcrumbCh is already full. A dropped diagnostic is not silent: it is
		// counted (droppedBreadcrumbs) and reported once room reappears (the next
		// Send call's flushDroppedBreadcrumbs, above).
		if pw.readout.breadcrumbCh != nil {
			select {
			case pw.readout.breadcrumbCh <- RowEvent{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbWireSendBufferFull, Debug: 1,
				NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Value: int32(v),
			}:
			default:
				pw.readout.droppedBreadcrumbs++
			}
		}
		return SendBufferFull
	}
}

// RecvTick is the non-blocking receive used by windowed nodes (In.PollRecv) on the
// DESTINATION node's own goroutine. Returns immediately: ok=false when the wire's
// own goroutine has not yet handed a delivered bead onto outCh. Also returns the
// tick the delivered bead actually landed on (deliveredBead.deliverTick, stamped
// by the wire's own goroutine at handoff time) — callers proving same-tick fan-out
// delivery must use this instead of re-reading a live clock after the fact.
func (pw *PacedWire) RecvTick() (int, int64, bool) {
	select {
	case db := <-pw.outCh:
		return db.val, db.deliverTick, true
	default:
		return 0, 0, false
	}
}

// Recv is RecvTick without the tick. Consumes on read (no separate Done step).
func (pw *PacedWire) Recv() (int, bool) {
	v, _, ok := pw.RecvTick()
	return v, ok
}

// ClearInFlight drops every bead this wire is carrying — the ones already crossing
// (inflight) AND the ones placed but not yet drained off inCh — so the wire is left
// as empty as a freshly built one. Nothing is delivered: a cleared bead never reaches
// the destination's outCh, which is the point. Beads already handed off to outCh are
// NOT touched here; they belong to the DESTINATION node now, and that node drains its
// own In (docs/beads-are-the-edge.md's ownership split).
//
// Same single-goroutine contract as DriveOneCycle/LiveBeadFractions, and for the same
// reason: inflight and inCh are owned by whichever goroutine drives this wire, which in
// production is the SOURCE node's own mover (nodeMover.run). Call it only from there —
// a node goroutine that wants its own outgoing wires cleared asks its mover
// (moveMsgKindBeadClear), it does not reach in here itself.
//
// The one caller today is the pair's RESET (nodes/PairNode): a reset means the
// straightening exchange is over, and beads still crossing would land a moment later and
// step the tilt straight back off zero — so returning the indices without emptying the
// bead edge does not actually stop anything.
func (pw *PacedWire) ClearInFlight() {
	for {
		select {
		case <-pw.inCh:
		default:
			pw.inflight = nil
			return
		}
	}
}

// DriveOneCycle is this wire's single per-cycle unit of work: drain newly
// Send-ed beads off inCh (stamping their placementTick from req.placementTick
// — the SENDING node's own clock reading, taken once at Send time, not read
// here), advance every in-flight bead due at tick by one position-step
// (emitting Position traces), and hand off any bead that has reached its
// delivery deadline onto outCh (non-blocking — a destination that hasn't
// drained yet simply leaves the bead retried next cycle, still at the FIFO
// head).
//
// In production this is called ONLY by the source node's own mover
// (nodeMover.run, node_mover.go), once per cycle for each of that node's
// outWires — the wire has no goroutine of its own; it is a passive struct
// stepped by whichever goroutine drains it. tick here paces stepAll's
// position-advance (the DRAINING side's own clock), which is a different
// concern from placementTick (the SENDING side's clock) above. It is exported
// so tests that build a bare PacedWire directly (no full loader topology,
// hence no nodeMover to drive it) can spawn their own driving goroutine that
// mimics production's per-cycle drive, exactly as StepOnceAt's callers used to.
func (pw *PacedWire) DriveOneCycle(ctx context.Context, tick int64) {
	if ctx.Err() != nil {
		return
	}
	pw.drainPlacements()
	pw.stepAll(tick)
}

// drainPlacements pops every placement currently queued on inCh (non-blocking)
// and appends each as a fresh in-flight bead, stamping placementTick from
// req.placementTick — the SENDING node's own clock reading, taken once by
// Send's caller at emission time (placeRequest's doc comment), not by this
// drain. Called only by the wire's driver (its source node's mover — see the
// PacedWire doc comment above; the wire itself has no goroutine).
//
// DRAIN-UNTIL-EMPTY, NOT AN ITERATION CAP (docs/planning/visual-editor/session-log.md Step 3, "the drain
// loops"). This is the canonical instance of a shape repeated at several
// other sites in this repo (nodes/wire/paced_wire.go's drainBreadcrumbEvents,
// nodes/gatecommon/gate.go's drainLatestReal, nodes/TimeStart/node.go and
// nodes/Time/node.go's mid-window observe loop, nodes/Wiring/edge_mover.go's
// and node_mover.go's per-cycle inbox drains — each of those points back
// here instead of repeating this paragraph). The reasoning that justifies
// EVERY one of them, and why none gets an arbitrary cap:
//
//  1. Each loop terminates the moment its channel reports empty (the
//     `default:` branch) — it never blocks waiting for more.
//  2. Every producer feeding one of these channels is itself bounded by a
//     DECLARED channel capacity (wireChanBufferSize here; moverInboxDepth
//     for the mover inboxes) — so no matter how busy the producer has been,
//     one drain call can pull at most that many items before hitting empty.
//     The loop is therefore transitively bounded by a number that already
//     has its own name elsewhere, not unbounded in practice.
//  3. Adding an iteration cap on top would not remove any risk — it would
//     ADD a new failure mode: a capped drain leaves items stranded in the
//     channel for a full extra cycle, which is worse than draining all of
//     them now. There is no capacity problem here to solve with a cap.
func (pw *PacedWire) drainPlacements() {
	for {
		select {
		case req := <-pw.inCh:
			pw.nextGen++
			pw.inflight = append(pw.inflight, inflightBead{
				val:           req.val,
				placementTick: float64(req.placementTick),
				// steps travels straight from the placement (beadPlacement.Steps,
				// copied from the sending Out's own Geom().Steps) — there is no
				// length to reconstruct here, unlike the retired arc/ms model,
				// because the sender already carries the bead-step count as an
				// integer all the way from PublishGeom (docs/bead-lattice.md
				// "The count").
				steps:   req.bp.Steps,
				seg:     WireSegment{Start: req.bp.Start, End: req.bp.End},
				node:    req.bp.Node,
				port:    req.bp.Port,
				streams: req.bp.streams(),
				gen:     pw.nextGen,
			})
			if len(pw.inflight) > maxInflightBeads {
				panic(fmt.Sprintf(
					"paced_wire: inflight exceeded %d beads on wire -> %s.%s; beads are being "+
						"placed faster than they cross and deliver. Two causes reach this: the "+
						"destination stopped draining outCh (FIFO-head delivery stalled), or the "+
						"source is placing faster than this wire can carry",
					maxInflightBeads, pw.Target, pw.TargetHandle))
			}
		default:
			return
		}
	}
}

// stepAll advances every in-flight bead due at tick by one position-step and
// attempts FIFO-head delivery for any that have reached their deadline. It
// processes beads head-first so an earlier bead's delivery in this same call can
// unblock a later bead's delivery within the same cycle — the same shape the old
// per-call gens-snapshot loop had; only this wire's own
// goroutine ever calls it.
func (pw *PacedWire) stepAll(tick int64) {
	nowTick := float64(tick)
	for i := 0; i < len(pw.inflight); {
		b := &pw.inflight[i]
		if !b.finalPending {
			if nowTick <= b.placementTick {
				i++
				continue
			}
			emit, pos, final := pw.advanceBead(b, nowTick)
			if emit && edgeBeadTraceEnabled && pw.readout.StreamsActive {
				pw.readout.appendPending(pendingWireEvent{
					kind: T.KindEdgeBead, value: pos.val,
					x: pos.x, y: pos.y, z: pos.z, t: pos.t, gen: pos.gen,
				}, pw.Target, pw.TargetHandle)
			}
			if !final {
				i++
				continue
			}
			b.finalPending = true
		}
		// Only the FIFO head (i==0) can deliver; a non-head bead simply waits.
		if i != 0 {
			i++
			continue
		}
		select {
		case pw.outCh <- deliveredBead{val: b.val, deliverTick: tick}:
			pw.emitArrive(arriveInfo{emit: b.streams, node: b.node, port: b.port, value: b.val, gen: b.gen})
			pw.inflight = pw.inflight[1:]
			if len(pw.inflight) == 0 {
				// inflight[1:] never re-slices the backing array; it only shrinks
				// the header while the array keeps growing with every
				// drainPlacements append. Once the slice is fully drained, drop
				// the reference so the (potentially large) backing array can be
				// collected instead of an append later reusing (and growing
				// further) a nil-length slice that still points at it. No
				// mid-slice compaction is attempted -- this is the one point
				// where a reset is free, because the slice is already empty.
				pw.inflight = nil
			}
			// Do not advance i: the slice shifted, so index 0 is now the next bead.
		default:
			// Destination hasn't drained outCh yet; retry the handoff next cycle.
			i++
		}
	}
}

// ReviseInFlightGeometry re-derives EVERY in-flight bead's remaining travel after a
// geometry edit (node-move) changed the edge (MODEL.md "Geometry and time"). It
// preserves each bead's FRACTIONAL progress t (its proportion along the wire), NOT
// the absolute distance covered: each bead stays at the same fraction t and the
// remaining time is recomputed from the NEW step count so UNIFORM PULSE SPEED holds —
// remaining = (1−t)·newSteps·dwell. driveOneCycle re-reads each bead's live
// steps/seg every cycle, so the new geometry takes effect without any relaunch.
// No-op when no bead is in flight.
//
// Called only by this wire's own goroutine (edgeMover.recomputeGeometry, itself
// running on the SAME goroutine as DriveOneCycle — see the PacedWire doc comment):
// tick is that goroutine's own clock reading, taken once per call. There is no
// second clock copy involved anymore, so the two-copy skew the old
// caller-pinned-tick contract had to tolerate cannot arise here.
func (pw *PacedWire) ReviseInFlightGeometry(tick int64, newSteps int, newSeg WireSegment) {
	if len(pw.inflight) == 0 {
		return
	}
	nowTick := float64(tick)
	for i := range pw.inflight {
		b := &pw.inflight[i]
		t := 0.0
		oldCross := pw.ticksToCross(b.steps)
		if oldCross > 0 {
			// elapsed ticks / old ticksToCross = fraction covered.
			t = (nowTick - b.placementTick) / oldCross
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
		}
		b.steps = newSteps
		b.seg = newSeg
		// Rebase placementTick so elapsed-since-placement maps to the same fraction t
		// on the NEW step count: remainingTicks = (1−t)·newSteps·dwell, so the covered
		// part is t·newSteps·dwell ticks ⇒ placementTick' = nowTick − t·(newSteps·dwell).
		// ticksToCross is 0 when dwell<=0, so this rebase is safe (a no-op shift)
		// even on a wire with no dwell configured; no separate guard needed.
		b.placementTick = nowTick - t*pw.ticksToCross(newSteps)
	}
}

// arriveInfo carries the source identity a delivery must echo on the arrive trace.
// emit is false for a bead that carried no position stream.
type arriveInfo struct {
	emit       bool
	node, port string
	value      int
	gen        uint64 // the delivered bead's per-wire id (renderer bead key)
}

// posEmitArgs holds the arguments for a tr.Position call, returned by
// advanceBead so the caller can emit it.
type posEmitArgs struct {
	node, port string
	val        int
	x, y, z, t float64
	gen        uint64
}

// emitArrive sends the traversal-complete trace for a delivered bead. Called by
// this wire's own goroutine (stepAll) right after the outCh handoff succeeds.
func (pw *PacedWire) emitArrive(ai arriveInfo) {
	if ai.emit && pw.readout.StreamsActive {
		pw.readout.appendPending(pendingWireEvent{kind: T.KindArrive, value: ai.value, gen: ai.gen},
			pw.Target, pw.TargetHandle)
	}
}

// advanceBead performs one cycle's work for the in-flight bead b at clock
// reading now (the scheduled tick time). Called only by this wire's own
// goroutine (stepAll).
//
// Returns:
//   - emit=true if a Position trace should be sent (tr.Position) for this call;
//     pos contains the arguments.
//   - final=true if the bead has reached or passed its deadline at now, meaning
//     the caller should attempt the FIFO-head delivery handoff.
func (pw *PacedWire) advanceBead(b *inflightBead, nowTick float64) (emit bool, pos posEmitArgs, final bool) {
	tr := pw.readout.Trace

	steps := b.steps
	seg := b.seg
	placementTick := b.placementTick
	stream := b.streams && tr != nil && steps > 0
	crossTicks := pw.ticksToCross(steps)

	deadline := placementTick + crossTicks

	target := nowTick
	if nowTick >= deadline {
		target = deadline
		final = true
	}

	if stream {
		// fractional progress t = elapsed ticks / ticksToCross (== steps
		// covered / steps, since both scale by the uniform per-bead dwell).
		t := 0.0
		if crossTicks > 0 {
			t = (target - placementTick) / crossTicks
		}
		if t > 1 {
			t = 1
		}
		p := lerp(seg.Start, seg.End, t)
		emit = true
		pos = posEmitArgs{
			node: b.node, port: b.port, val: b.val,
			x: p.X, y: p.Y, z: p.Z, t: t, gen: b.gen,
		}
	}
	return
}
