package beadanimation

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
