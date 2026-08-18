package wire

import (
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"
)

var edgeBeadTraceEnabled = os.Getenv("WIREFOLD_EDGE_BEAD_TRACE") == "1"

const wireChanBufferSize = 4096

type deliveredBead struct {
	val         int
	deliverTick int64
}

type PacedWire struct {
	inCh  chan placeRequest
	outCh chan deliveredBead

	kindToAnimClearCh chan struct{}

	inflight []inflightBead

	nextGen uint64

	dwell float64

	Owner string
	Edge  string

	Target       string
	TargetHandle string

	readout wireReadout

	rev revisionSlot
}

const maxInflightBeads = wireChanBufferSize

func NewPacedWire(steps int, dwellTicks float64) *PacedWire {
	return &PacedWire{
		dwell: dwellTicks,
		inCh:  make(chan placeRequest, wireChanBufferSize),
		outCh: make(chan deliveredBead, wireChanBufferSize),

		kindToAnimClearCh: make(chan struct{}, 1),

		readout: wireReadout{breadcrumbCh: make(chan rowevent.RowEvent, 4)},
		rev:     newRevisionSlot(),
	}
}
