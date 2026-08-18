package rowevent

type RowEvent struct {
	Kind                                                             string
	NodeRow, PortRow, TargetRow, TargetPortRow, EdgeRow, Slot, Value int32
	Bead                                                             uint64

	BeadSteps  float64
	X, Y, Z, F float64

	Label uint8
	Debug uint8
	Text  string
}

type EventSink interface {
	WriteEvents(events []RowEvent)
	NodeRowOf() int32
}
