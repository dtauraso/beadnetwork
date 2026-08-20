package interior

import "github.com/dtauraso/wirefold/src/Node/rowevent"

const SlotsPerNode = 4

type Emitter struct {
	mailbox *Mailbox
	nodeRow int32
}

func NewEmitter(mailbox *Mailbox, nodeRow int32) *Emitter {
	return &Emitter{mailbox: mailbox, nodeRow: nodeRow}
}

func (e *Emitter) NodeRowOf() int32 {
	if e == nil {
		return -1
	}
	return e.nodeRow
}

func (e *Emitter) WriteEvents(events []rowevent.RowEvent) {
	if e == nil {
		return
	}
	e.mailbox.Send(Snapshot{EventsOnly: true, Events: events})
}

func (e *Emitter) write(present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) {
	if e == nil {
		return
	}
	e.mailbox.Send(Snapshot{
		Present: present, Value: value,
		Ox: ox, Oy: oy, Oz: oz,
		Events: events,
	})
}
