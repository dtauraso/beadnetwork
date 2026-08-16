package wire

import (
	"context"
	"fmt"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/spatial"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

func (pw *PacedWire) DriveOneCycle(ctx context.Context, tick int64) {
	if ctx.Err() != nil {
		return
	}
	pw.applyRevision(tick)
	pw.drainPlacements(tick)
	pw.stepAll(tick)
}

func (pw *PacedWire) drainPlacements(tick int64) {
	for {
		select {
		case req := <-pw.inCh:
			pw.nextGen++
			pw.inflight = append(pw.inflight, inflightBead{
				val:           req.val,
				placementTick: float64(tick),

				steps:   req.bp.Steps,
				seg:     spatial.WireSegment{Start: req.bp.Start, End: req.bp.End},
				node:    req.bp.Node,
				port:    req.bp.Port,
				streams: req.bp.streams(),
				gen:     pw.nextGen,
			})
			if len(pw.inflight) > maxInflightBeads {
				panic(fmt.Sprintf(
					"paced_wire: inflight exceeded %d beads on edge %q owned by node %s (-> %s.%s); "+
						"beads are being placed faster than they cross and deliver. Two causes reach "+
						"this: the destination stopped draining outCh (FIFO-head delivery stalled), "+
						"or the source is placing faster than this wire can carry",
					maxInflightBeads, pw.Edge, pw.Owner, pw.Target, pw.TargetHandle))
			}
		default:
			return
		}
	}
}

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
				}, pw.Owner, pw.Edge)
			}
			if !final {
				i++
				continue
			}
			b.finalPending = true
		}

		if i != 0 {
			i++
			continue
		}
		select {
		case pw.outCh <- deliveredBead{val: b.val, deliverTick: tick}:
			pw.emitArrive(arriveInfo{emit: b.streams, node: b.node, port: b.port, value: b.val, gen: b.gen})
			pw.inflight = pw.inflight[1:]
			if len(pw.inflight) == 0 {

				pw.inflight = nil
			}

		default:

			i++
		}
	}
}

func (pw *PacedWire) ReviseInFlightGeometry(tick int64, newSteps int, newSeg spatial.WireSegment) {
	if len(pw.inflight) == 0 {
		return
	}
	nowTick := float64(tick)
	for i := range pw.inflight {
		b := &pw.inflight[i]

		oldCross := pw.ticksToCross(b.steps)
		t := lattice.BeadFraction(nowTick, b.placementTick, oldCross)
		b.steps = newSteps
		b.seg = newSeg

		b.placementTick = nowTick - t*pw.ticksToCross(newSteps)
	}
}
