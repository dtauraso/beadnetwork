package wire

type EventSink interface {
	WriteEvents(events []RowEvent)
	NodeRowOf() int32
}
