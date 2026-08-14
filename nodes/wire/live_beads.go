package wire

import (
	"github.com/dtauraso/wirefold/nodes/spatial"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

type LiveBeadRow struct {
	Val     int
	X, Y, Z float64
	Gen     uint64

	Steps int
	Age   float64
}

func (pw *PacedWire) LiveBeadRows(tick int64) []LiveBeadRow {
	nowTick := float64(tick)
	rows := make([]LiveBeadRow, 0, len(pw.inflight))
	for i := range pw.inflight {
		b := &pw.inflight[i]
		if !b.streams {
			continue
		}
		crossTicks := pw.ticksToCross(b.steps)
		t := lattice.BeadFraction(nowTick, b.placementTick, crossTicks)
		p := spatial.Lerp(b.seg.Start, b.seg.End, t)
		rows = append(rows, LiveBeadRow{
			Val: b.val, X: p.X, Y: p.Y, Z: p.Z, Gen: b.gen,
			Steps: b.steps, Age: nowTick - b.placementTick,
		})
	}
	return rows
}
