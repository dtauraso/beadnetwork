package wire

import (
	"os"
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

	inflight []inflightBead

	nextGen uint64

	dwell float64

	Target       string
	TargetHandle string

	readout wireReadout
}

const maxInflightBeads = wireChanBufferSize

func NewPacedWire(steps int, dwellTicks float64) *PacedWire {
	return &PacedWire{
		dwell:   dwellTicks,
		inCh:    make(chan placeRequest, wireChanBufferSize),
		outCh:   make(chan deliveredBead, wireChanBufferSize),
		readout: wireReadout{breadcrumbCh: make(chan RowEvent, 4)},
	}
}
