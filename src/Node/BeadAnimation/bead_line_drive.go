package beadanimation

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/src/spatial"
)

func (bl *BeadLine) DriveOneStep(ctx context.Context, tick int64) {
	if ctx.Err() != nil {
		return
	}
	bl.applyClear()
	bl.drainPlacements()
	bl.stepAll(tick)
}

func (bl *BeadLine) drainPlacements() {
	for {
		select {
		case req := <-bl.inCh:
			bl.nextGen++
			bl.inflight = append(bl.inflight, inflightBead{
				val:     req.val,
				slot:    0,
				seg:     spatial.Segment{Start: req.bp.Start, End: req.bp.End},
				steps:   req.bp.Steps,
				slotR:   req.bp.SlotR,
				node:    req.bp.Node,
				port:    req.bp.Port,
				streams: req.bp.streams(),
				gen:     bl.nextGen,
			})
			if len(bl.inflight) > maxInflightBeads {
				panic(fmt.Sprintf(
					"BeadAnimation: inflight exceeded %d beads on edge %q owned by node %s (-> %s.%s); "+
						"beads are being placed faster than they cross and deliver. Two causes reach "+
						"this: the destination stopped draining outCh (FIFO-head delivery stalled), "+
						"or the source is placing faster than this run can carry",
					maxInflightBeads, bl.Edge, bl.Owner, bl.Target, bl.TargetHandle))
			}
		default:
			return
		}
	}
}

func (bl *BeadLine) stepAll(tick int64) {
	for i := range bl.inflight {
		b := &bl.inflight[i]
		bl.advance(b)
	}

	for len(bl.inflight) > 0 {
		b := &bl.inflight[0]
		if !bl.arrived(b) {
			return
		}
		select {
		case bl.outCh <- deliveredBead{val: b.val, deliverTick: tick}:
			bl.emitArrive(arriveInfo{emit: b.streams, node: b.node, port: b.port, value: b.val, gen: b.gen})
			bl.inflight = bl.inflight[1:]
			if len(bl.inflight) == 0 {
				bl.inflight = nil
			}
		default:
			return
		}
	}
}
