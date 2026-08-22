package interior

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

func (e *Emitter) WriteEvents(events []RowEvent) {
	if e == nil {
		return
	}
	e.mailbox.Send(Snapshot{EventsOnly: true, Events: events})
}

func (e *Emitter) Recv(portRow, value int32) {
	e.WriteEvents([]RowEvent{{
		Kind: KindRecv, NodeRow: e.NodeRowOf(), PortRow: portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: value,
	}})
}

func (e *Emitter) Send(portRow, targetRow, targetPortRow, value int32, steps float64) {
	e.WriteEvents([]RowEvent{{
		Kind: KindSend, NodeRow: e.NodeRowOf(), PortRow: portRow,
		TargetRow: targetRow, TargetPortRow: targetPortRow, EdgeRow: -1,
		Value: value, BeadSteps: steps,
	}})
}

func (e *Emitter) Fire() {
	e.WriteEvents([]RowEvent{{
		Kind: KindFire, NodeRow: e.NodeRowOf(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}})
}

func (e *Emitter) Breadcrumb(label string, portRow int32) {
	e.WriteEvents([]RowEvent{{
		Kind: KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: e.NodeRowOf(), PortRow: portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
	}})
}

func (e *Emitter) write(present []uint8, value []int32, ox, oy, oz []float32, events []RowEvent) {
	if e == nil {
		return
	}
	e.mailbox.Send(Snapshot{
		Present: present, Value: value,
		Ox: ox, Oy: oy, Oz: oz,
		Events: events,
	})
}
