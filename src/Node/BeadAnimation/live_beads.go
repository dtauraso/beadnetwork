package beadanimation

type LiveBeadRow struct {
	Val     int
	X, Y, Z float64
	Gen     uint64

	Steps int
	Slot  int
}

func (bl *BeadLine) LiveBeadRows() []LiveBeadRow {
	rows := make([]LiveBeadRow, 0, len(bl.inflight))
	for i := range bl.inflight {
		b := &bl.inflight[i]
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
