// paced_wire.go — ONE JOB: what a wire IS. The PacedWire struct itself (its two
// channels, its owned in-flight queue, its dwell and its readout), its constructor,
// and the package-level constants and small value types that describe a wire's
// capacity and units. What a wire DOES is split by job across its siblings:
// bead_placement.go (what a bead carries), arrival.go (when it lands),
// paced_wire_send.go (the channel endpoints), paced_wire_drive.go (the per-cycle
// step) and bead_advance.go (one bead's step).

package wire

import (
	"os"
)

// edgeBeadTraceEnabled gates whether stepAll appends a T.KindEdgeBead pendingWireEvent
// for every in-flight bead every tick. It is read ONCE at process startup from the
// WIREFOLD_EDGE_BEAD_TRACE env var (same "one env var, read once before any wire
// goroutine exists" shape as Buffer/streamframe/stream_fds.go's WIREFOLD_STREAM_FDS) -- never
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

// PulseSpeedWuPerMs / PulseSpeedWuPerTick — the uniform pulse speed, in world-units
// per millisecond and per tick — now live in nodes/wire/lattice (lattice.go), not
// here: DwellTicksPerBead (also in that package) derives from both BeadStepR and
// PulseSpeedWuPerTick, and nodes/wire itself needs DwellTicksPerBead (out_port.go's
// SimLatencyMs), so the constants those two depend on cannot live in package wire
// without a wire<->lattice import cycle.

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
	// dwell is this wire's own ticks-per-bead-step (docs/bead-model/bead-lattice.md
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
// That one-call-site invariant is enforced by tools/network/beads/check-uniform-pulse-speed.sh. Do not
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
