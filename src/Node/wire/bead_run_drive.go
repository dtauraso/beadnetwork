package wire

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/src/Node/spatial"
)

func (pw *BeadRun) DriveOneStep(ctx context.Context, tick int64) {
	if ctx.Err() != nil {
		return
	}
	pw.applyClear()
	pw.drainPlacements()
	pw.stepAll(tick)
}

func (pw *BeadRun) drainPlacements() {
	for {
		select {
		case req := <-pw.inCh:
			pw.nextGen++
			pw.inflight = append(pw.inflight, inflightBead{
				val:     req.val,
				slot:    0,
				seg:     spatial.Segment{Start: req.bp.Start, End: req.bp.End},
				steps:   req.bp.Steps,
				slotR:   req.bp.SlotR,
				node:    req.bp.Node,
				port:    req.bp.Port,
				streams: req.bp.streams(),
				gen:     pw.nextGen,
			})
			if len(pw.inflight) > maxInflightBeads {
				panic(fmt.Sprintf(
					"bead_run: inflight exceeded %d beads on edge %q owned by node %s (-> %s.%s); "+
						"beads are being placed faster than they cross and deliver. Two causes reach "+
						"this: the destination stopped draining outCh (FIFO-head delivery stalled), "+
						"or the source is placing faster than this run can carry",
					maxInflightBeads, pw.Edge, pw.Owner, pw.Target, pw.TargetHandle))
			}
		default:
			return
		}
	}
}

func (pw *BeadRun) stepAll(tick int64) {
	for i := range pw.inflight {
		b := &pw.inflight[i]
		pw.advance(b)
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
