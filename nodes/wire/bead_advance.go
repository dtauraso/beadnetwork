package wire

import (
	T "github.com/dtauraso/wirefold/Trace"
)

type arriveInfo struct {
	emit       bool
	node, port string
	value      int
	gen        uint64
}

func (pw *PacedWire) emitArrive(ai arriveInfo) {
	if ai.emit && pw.readout.StreamsActive {
		pw.readout.appendPending(pendingWireEvent{kind: T.KindArrive, value: ai.value, gen: ai.gen},
			pw.Owner, pw.Edge)
	}
}

func (pw *PacedWire) advance(b *inflightBead) {
	b.slot++
}
