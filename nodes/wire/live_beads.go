package wire

import (
	"github.com/dtauraso/wirefold/nodes/spatial"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

type LiveBeadRow struct {
	Val     int
	X, Y, Z float64
	Gen     uint64

	// Steps and Age are what decided this bead's fraction: the crossing it
	// was placed for, and how far into it the bead already was when this row
	// was taken. They ride along so a bead in the wrong place can be read
	// back to whichever of the two was wrong.
	Steps int
	Age   float64
}

type LiveBeadProgress struct {
	T   float64
	Val int

	Steps int
}

func (pw *PacedWire) LiveBeadFractions(tick int64) []LiveBeadProgress {
	nowTick := float64(tick)
	out := make([]LiveBeadProgress, 0, len(pw.inflight))
	for i := range pw.inflight {
		b := &pw.inflight[i]
		crossTicks := pw.ticksToCross(b.steps)
		if crossTicks <= 0 {
			continue
		}
		t := lattice.BeadFraction(nowTick, b.placementTick, crossTicks)
		out = append(out, LiveBeadProgress{T: t, Val: b.val, Steps: b.steps})
	}
	return out
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
