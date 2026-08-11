// paced_wire_drive.go — ONE JOB: the per-cycle stepping of the wire by whichever
// goroutine owns it (its SOURCE NODE's mover). DriveOneCycle and its two halves —
// drainPlacements (in-channel -> inflight) and stepAll (advance + FIFO-head handoff
// onto the out-channel) — plus ReviseInFlightGeometry, the other thing that same
// owning goroutine does to the same in-flight set when a node-move reshapes the
// edge. The per-bead math each step runs is bead_advance.go.

package wire

import (
	"context"
	"fmt"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

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
// PacedWire doc comment in paced_wire.go; the wire itself has no goroutine).
//
// DRAIN-UNTIL-EMPTY, NOT AN ITERATION CAP (docs/planning/visual-editor/session-log.md Step 3, "the drain
// loops"). This is the canonical instance of a shape repeated at several
// other sites in this repo (nodes/wire/wire_readout.go's drainBreadcrumbEvents,
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
				// integer all the way from PublishGeom (docs/bead-model/bead-lattice.md
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
		// elapsed ticks / old ticksToCross = fraction covered.
		oldCross := pw.ticksToCross(b.steps)
		t := lattice.BeadFraction(nowTick, b.placementTick, oldCross)
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
