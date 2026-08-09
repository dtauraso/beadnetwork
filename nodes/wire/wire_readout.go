// wire_readout.go — the REPORTING half a PacedWire composes.
//
// PacedWire is a passive delay queue (MODEL.md "Wire (PacedWire)"), but roughly half its
// struct was never about transport at all: the pending Position/Arrive buffer the renderer
// drains, the Trace handle, the debug-breadcrumb channel and its drop counter. Those say
// nothing about when a bead lands; they exist so a frame can be written. This file gives
// that concern a NAMED type; paced_wire.go keeps the queue itself (arrival.go its arrival
// math, paced_wire_drive.go the stepping of it).
//
// Same pattern nodeGeometry already follows (nodes/Wiring/node_geometry_parts.go): a NAMED
// sub-object accessed explicitly (pw.readout.pending), never Go embedding — embedding would
// keep the flat namespace and hide the owner.
//
// OWNERSHIP IS UNCHANGED BY THE SPLIT. Every field here is owned by the same single
// goroutine that owned it as a loose PacedWire field — pending by this wire's own driving
// goroutine (its SOURCE NODE's mover), droppedBreadcrumbs by that same source goroutine at
// its Send call site, StreamsActive/Trace written once at wiring time before that goroutine
// launches. No lock and no atomic appears here, and none may be added: ownership replaces
// locking (MODEL.md, memory/feedback_no_atomics_are_defects.md,
// tools/check-no-network-locks.sh with an EMPTY allowlist).
package wire

import (
	"fmt"

	T "github.com/dtauraso/wirefold/Trace"
)

// wireReadout is a wire's reporting apparatus: what it tells the renderer and the debug
// log, kept separate from the delay queue that actually carries beads. It holds no
// transport state — nothing here decides when a bead arrives — and the queue holds no
// reporting state.
type wireReadout struct {
	// Trace is injected by the loader; used for breadcrumb diagnostics only.
	Trace *T.Trace

	// StreamsActive reports whether a real consumer is wired for this edge's
	// pending-event buffer (Buffer/stream_fds.go's per-edge fd, via
	// streamWiring.setEdgeStreams — see edgeMover.streamOut). Set EXACTLY ONCE,
	// through PacedWire.SetStreamsActive, at wiring time BEFORE this edge's mover
	// goroutine launches (the same "wire before launch, read-only afterward"
	// ordering documented on streamWiring.interiorOuts and move_streams.go —
	// stream_wiring.go:28), and never written again afterward — so, like Trace
	// above, it needs no lock/atomic (memory/feedback_no_atomics_are_defects.md).
	// Default false (bare test construction, or a real edge with no fd entry in
	// WIREFOLD_STREAM_FDS): pending MUST NOT accumulate with nothing to ever
	// drain it — see emitArrive and stepAll's KindEdgeBead append, both gated on
	// this field. This is deliberately NOT beadPlacement.streams() (bp.Node !=
	// "") — that reports whether ONE PLACEMENT carries position-stream
	// geometry, a per-bead property completely independent of whether any
	// stream consumer exists for this wire at all; conflating the two was the
	// root cause of the confirmed unbounded-pending-growth bug documented in
	// docs/planning/visual-editor/session-log.md (streamOut nil -> 40
	// deliveries left all 40 queued forever with nothing ever calling
	// DrainPendingEvents).
	StreamsActive bool

	// pending buffers this wire's OWN Position/Arrive events since the last drain
	// (memory/feedback_no_single_writer_bridge.md): appended only by stepAll (this
	// wire's own goroutine, via edgeMover.run's DriveOneCycle call) and drained only by
	// edgeMover.writeStreamFrame — the SAME goroutine on both ends (edgeMover.run calls
	// DriveOneCycle then writeStreamFrame back to back, every cycle).
	pending []pendingWireEvent

	// breadcrumbCh carries this wire's "wire-send-buffer-full" DEBUG BREADCRUMB
	// (Send below) from the SOURCE node's own goroutine (a DIFFERENT goroutine than
	// this wire's own — unlike pending above, which only the wire's own goroutine
	// ever touches) over to this wire's own goroutine, which drains it in
	// edgeMover.writeStreamFrame alongside drainPendingEvents. Non-blocking send,
	// same "no delivery guarantee, may drop a diagnostic" shape as abcDragCh
	// (view_stream.go) — this should never fire under realistic load anyway.
	breadcrumbCh chan RowEvent

	// droppedBreadcrumbs counts breadcrumbCh sends Send's non-blocking send
	// dropped (channel full) since the last flushDroppedBreadcrumbs report.
	// breadcrumbCh has exactly one producer call site (Send, called only from
	// this wire's fixed SOURCE node's own goroutine — nodes/wire/out_port.go's
	// one call site), so this field is owned EXCLUSIVELY by that same source
	// goroutine, never touched by this wire's own goroutine — a different,
	// but equally single-owner, contract than pending/inflight above
	// (memory/feedback_no_atomics_are_defects.md). The cap-4 breadcrumbCh is
	// the tightest queue in the system and drops are non-blocking by design
	// (breadcrumbs must never backpressure the network); this counter exists
	// so a drop is reported once room reappears instead of vanishing silently
	// — see flushDroppedBreadcrumbs.
	droppedBreadcrumbs int
}

// SetTrace injects the loader's Trace handle onto this wire's readout (build.go's
// wiring pass, before this edge's mover goroutine launches).
func (pw *PacedWire) SetTrace(tr *T.Trace) { pw.readout.Trace = tr }

// SetStreamsActive marks that a real per-edge stream consumer is wired for this wire —
// the single external write of readout.StreamsActive, whose doc comment carries the
// "exactly once, at wiring time, before the goroutine launches" contract.
func (pw *PacedWire) SetStreamsActive(active bool) { pw.readout.StreamsActive = active }

// flushDroppedBreadcrumbs reports (as a T.KindBreadcrumb row, Label
// BreadcrumbWireBreadcrumbsDropped, Value = the dropped count) and clears any
// breadcrumbCh drops recorded since the last flush, IF breadcrumbCh currently
// has room. Called at the top of every Send call — the same single
// caller/owner as droppedBreadcrumbs itself (this wire's fixed SOURCE node's
// own goroutine) — so a run of drops is eventually surfaced instead of lost
// for good. Non-blocking: if breadcrumbCh is still full, the count is left
// untouched and retried on the next Send call. A cheap no-op when nothing has
// been dropped (the overwhelming common case).
func (r *wireReadout) flushDroppedBreadcrumbs() {
	if r.breadcrumbCh == nil || r.droppedBreadcrumbs == 0 {
		return
	}
	select {
	case r.breadcrumbCh <- RowEvent{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbWireBreadcrumbsDropped, Debug: 1,
		NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(r.droppedBreadcrumbs),
	}:
		r.droppedBreadcrumbs = 0
	default:
	}
}

// drainBreadcrumbEvents returns every breadcrumb RowEvent buffered on breadcrumbCh
// since the last call (non-blocking drain). Safe to call only from this wire's own
// goroutine (edgeMover.writeStreamFrame), same contract as drainPendingEvents.
//
// Drain-until-empty, transitively bounded by breadcrumbCh's declared capacity (4) —
// no iteration cap; see drainPlacements's doc comment (paced_wire_drive.go) for the full
// reasoning shared by every drain-until-empty loop in this repo.
func (r *wireReadout) drainBreadcrumbEvents() []RowEvent {
	if r.breadcrumbCh == nil {
		return nil
	}
	var out []RowEvent
	for {
		select {
		case ev := <-r.breadcrumbCh:
			out = append(out, ev)
		default:
			return out
		}
	}
}

// DrainBreadcrumbEvents is drainBreadcrumbEvents' exported entry point for callers in
// another package (edgeMover.writeStreamFrame in nodes/Wiring).
func (pw *PacedWire) DrainBreadcrumbEvents() []RowEvent {
	return pw.readout.drainBreadcrumbEvents()
}

// pendingWireEvent is one raw Position/Arrive tuple recorded by stepAll, awaiting
// row-resolution + packing by edgeMover.writeStreamFrame (drainPendingEvents).
type pendingWireEvent struct {
	kind       string
	value      int
	x, y, z, t float64
	gen        uint64
}

// maxPendingEvents is the declared upper bound on len(pw.readout.pending) between drains.
// With a stream consumer wired (StreamsActive), pending is drained EVERY cycle by
// edgeMover.writeStreamFrame -> DrainPendingEvents (the same goroutine, back to
// back with the append side — see pending's own doc comment above), so its true
// maximum is one cycle's production. Exceeding that does not mean "busy"; with
// the drain running, it can only mean the drain has stopped running — the exact
// bug this branch's inventory found and 93d2e9b6 fixed (streamOut nil left
// pending accumulating forever with nothing ever calling DrainPendingEvents).
//
// The number itself is a GENEROUS ceiling, not a tight derivation, and that
// caveat is deliberate: pw.inflight (the source of every pending append) has no
// declared maximum of its own yet (docs/planning/visual-editor/session-log.md Step 3), so "one cycle's
// production" cannot be proven exactly today. wireChanBufferSize already
// ceilings how many beads can ever be outstanding on this wire via inCh, so
// reusing it here gives a value that is definitely wrong to reach rather than
// one derived airtight from a bounded inflight.
const maxPendingEvents = wireChanBufferSize

// appendPending appends ev to r.pending and asserts the result has not
// exceeded maxPendingEvents. It is the ONLY way pending is ever appended to
// (stepAll's KindEdgeBead append and emitArrive's KindArrive append both call
// this instead of appending directly) so the bound is checked at every mutation
// site, not just some. target/targetHandle are the owning wire's destination
// routing keys, passed in only to name the offending wire in the panic.
//
// Panics rather than dropping or growing further: this is an INVARIANT check on
// this wire's own goroutine, not a capacity policy (docs/planning/visual-editor/session-log.md Step 1) —
// dropping would hide a broken drain (the exact failure this bound exists to
// catch), and backpressure would couple this wire's pacing to its owner's frame
// rate. Reaching this bound is a code bug, never ordinary traffic, the same
// "can only break via a code bug" shape as wire.Register's and build.go's
// panics. The convention is stated in MODEL.md "Assertions" and enforced by
// tools/check-panic-message.sh.
func (r *wireReadout) appendPending(ev pendingWireEvent, target, targetHandle string) {
	r.pending = append(r.pending, ev)
	if len(r.pending) > maxPendingEvents {
		panic(fmt.Sprintf(
			"paced_wire: pending exceeded %d events on wire -> %s.%s; the per-cycle drain "+
				"(edgeMover.writeStreamFrame -> DrainPendingEvents) is not running",
			maxPendingEvents, target, targetHandle))
	}
}

// drainPendingEvents returns every pendingWireEvent recorded since the last call and
// clears the buffer. Safe to call only from this wire's own goroutine.
func (r *wireReadout) drainPendingEvents() []pendingWireEvent {
	if len(r.pending) == 0 {
		return nil
	}
	out := r.pending
	r.pending = nil
	return out
}

// PendingWireEvent is pendingWireEvent's exported mirror, returned by
// DrainPendingEvents for callers in another package (edgeMover.writeStreamFrame in
// nodes/Wiring) that need to row-resolve and pack these events but cannot name the
// unexported pendingWireEvent type.
type PendingWireEvent struct {
	Kind       string
	Value      int
	X, Y, Z, T float64
	Gen        uint64
}

// DrainPendingEvents is drainPendingEvents' exported entry point, converting each
// internal pendingWireEvent to the exported PendingWireEvent shape.
func (pw *PacedWire) DrainPendingEvents() []PendingWireEvent {
	internal := pw.readout.drainPendingEvents()
	if internal == nil {
		return nil
	}
	out := make([]PendingWireEvent, len(internal))
	for i, pe := range internal {
		out[i] = PendingWireEvent{Kind: pe.kind, Value: pe.value, X: pe.x, Y: pe.y, Z: pe.z, T: pe.t, Gen: pe.gen}
	}
	return out
}
