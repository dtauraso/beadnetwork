package bead

type LiveBeadRow struct {
	Val     int
	X, Y, Z float64
	Gen     uint64

	Steps int
	Slot  int
}

func (pw *BeadRun) LiveBeadRows() []LiveBeadRow {
	rows := make([]LiveBeadRow, 0, len(pw.inflight))
	for i := range pw.inflight {
		b := &pw.inflight[i]
		if !b.streams {
			continue
		}
		p := b.pos()
		rows = append(rows, LiveBeadRow{
			Val: b.val, X: p.X, Y: p.Y, Z: p.Z, Gen: b.gen,
			Steps: b.steps, Slot: b.slot,
		})
	}
	return rows
}
