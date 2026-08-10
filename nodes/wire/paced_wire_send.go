// paced_wire_send.go — ONE JOB: the wire's two channel ENDPOINTS, i.e. everything
// a node calls on a wire it does not drive-step. The source node's Send onto the
// in-channel (and the outcome type that keeps a transient buffer-full from reading
// as terminal), the destination node's RecvTick/Recv off the out-channel, and
// ClearInFlight, which empties both the in-channel and the in-flight queue. The
// per-cycle stepping between those two ends is paced_wire_drive.go.

package wire

import (
	T "github.com/dtauraso/wirefold/Trace"
)

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
// own In (docs/bead-model/beads-are-the-edge.md's ownership split).
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
