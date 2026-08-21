package beadanimation

import (
	T "github.com/dtauraso/wirefold/src/Trace"
)

type arriveInfo struct {
	emit       bool
	node, port string
	value      int
	gen        uint64
}

func (bl *BeadLine) emitArrive(ai arriveInfo) {
	if ai.emit && bl.readout.StreamsActive {
		bl.readout.appendPending(pendingBeadEvent{kind: T.KindArrive, value: ai.value, gen: ai.gen},
			bl.Owner, bl.Edge)
	}
}

func (bl *BeadLine) advance(b *inflightBead) {
	b.slot++
}
