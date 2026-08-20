package wire

import (
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

type arriveInfo struct {
	emit       bool
	node, port string
	value      int
	gen        uint64
}

func (pw *BeadRun) emitArrive(ai arriveInfo) {
	if ai.emit && pw.readout.StreamsActive {
		pw.readout.appendPending(pendingBeadEvent{kind: B.KindArrive, value: ai.value, gen: ai.gen},
			pw.Owner, pw.Edge)
	}
}

func (pw *BeadRun) advance(b *inflightBead) {
	b.slot++
}
