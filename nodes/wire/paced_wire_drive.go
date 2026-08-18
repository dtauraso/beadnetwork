package wire

import (
	"context"
	"fmt"
	"math"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func (pw *PacedWire) DriveOneStep(ctx context.Context, tick int64) {
	if ctx.Err() != nil {
		return
	}
	pw.applyClear()
	pw.applyRevision()
	pw.drainPlacements()
	pw.stepAll(tick)
}

func (pw *PacedWire) drainPlacements() {
	for {
		select {
		case req := <-pw.inCh:
			pw.nextGen++
			pw.inflight = append(pw.inflight, inflightBead{
				val:     req.val,
				slot:    0,
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
	for i := range pw.inflight {
		b := &pw.inflight[i]
		pw.advance(b)
		if edgeBeadTraceEnabled && pw.readout.StreamsActive && b.streams {
			p := b.pos()
			pw.readout.appendPending(pendingWireEvent{
				kind: T.KindEdgeBead, value: b.val,
				x: p.X, y: p.Y, z: p.Z, t: float64(b.slot), gen: b.gen,
			}, pw.Owner, pw.Edge)
		}
	}

	for len(pw.inflight) > 0 {
		b := &pw.inflight[0]
		if !pw.arrived(b) {
			return
		}
		select {
		case pw.outCh <- deliveredBead{val: b.val, deliverTick: tick}:
			pw.emitArrive(arriveInfo{emit: b.streams, node: b.node, port: b.port, value: b.val, gen: b.gen})
			pw.inflight = pw.inflight[1:]
			if len(pw.inflight) == 0 {
				pw.inflight = nil
			}
		default:
			return
		}
	}
}

func (pw *PacedWire) slotsPerBead() int {
	n := int(math.Round(pw.dwell))
	if n < 1 {
		return 1
	}
	return n
}

func (pw *PacedWire) ReviseGeometry(newSteps int, newSeg spatial.WireSegment) {
	for i := range pw.inflight {
		b := &pw.inflight[i]
		b.steps = newSteps
		b.seg = newSeg
	}
}
