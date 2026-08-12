package wire

type RowEvent struct {
	Kind                                                             string
	NodeRow, PortRow, TargetRow, TargetPortRow, EdgeRow, Slot, Value int32
	Bead                                                             uint64

	BeadSteps                float64
	SimLatencyMs, X, Y, Z, F float64

	Label uint8
	Debug uint8
	Text  string
}
