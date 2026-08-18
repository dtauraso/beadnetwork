package bead

import (
	T "github.com/dtauraso/wirefold/Trace"
)

type arriveInfo struct {
	emit       bool
	node, port string
	value      int
	gen        uint64
}

func (pw *BeadRun) emitArrive(ai arriveInfo) {
	if ai.emit && pw.readout.StreamsActive {
		pw.readout.appendPending(pendingBeadEvent{kind: T.KindArrive, value: ai.value, gen: ai.gen},
			pw.Owner, pw.Edge)
	}
}

func (pw *BeadRun) advance(b *inflightBead) {
	b.slot++
}
