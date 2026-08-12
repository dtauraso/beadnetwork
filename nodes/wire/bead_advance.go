package wire

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

type arriveInfo struct {
	emit       bool
	node, port string
	value      int
	gen        uint64
}

type posEmitArgs struct {
	node, port string
	val        int
	x, y, z, t float64
	gen        uint64
}

func (pw *PacedWire) emitArrive(ai arriveInfo) {
	if ai.emit && pw.readout.StreamsActive {
		pw.readout.appendPending(pendingWireEvent{kind: T.KindArrive, value: ai.value, gen: ai.gen},
			pw.Target, pw.TargetHandle)
	}
}

func (pw *PacedWire) advanceBead(b *inflightBead, nowTick float64) (emit bool, pos posEmitArgs, final bool) {
	tr := pw.readout.Trace

	steps := b.steps
	seg := b.seg
	placementTick := b.placementTick
	stream := b.streams && tr != nil && steps > 0
	crossTicks := pw.ticksToCross(steps)

	deadline := placementTick + crossTicks
	final = nowTick >= deadline

	if stream {

		t := lattice.BeadFraction(nowTick, placementTick, crossTicks)
		p := lerp(seg.Start, seg.End, t)
		emit = true
		pos = posEmitArgs{
			node: b.node, port: b.port, val: b.val,
			x: p.X, y: p.Y, z: p.Z, t: t, gen: b.gen,
		}
	}
	return
}
