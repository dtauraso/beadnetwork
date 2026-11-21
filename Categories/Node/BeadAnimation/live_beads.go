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

const dragInboxDepth = 8

type Beads struct {
	dragToAnim chan bool
}

func (nb *Beads) PostBeadDrag(start bool) {
	if nb.dragToAnim == nil {
		nb.dragToAnim = make(chan bool, dragInboxDepth)
	}
	select {
	case nb.dragToAnim <- start:
	default:
		panic("Beads: bead-drag inbox full — a drag start/end is one event per human drag, " +
			"so a full queue means this node's animation goroutine has stopped running, " +
			"not that drags arrived too fast")
	}
}

func (nb *Beads) ApplyBeadDrag() {
	for {
		select {
		case <-nb.dragToAnim:
		default:
			return
		}
	}
}
