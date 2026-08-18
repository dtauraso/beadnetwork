package bead

import (
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"
)

var edgeBeadTraceEnabled = os.Getenv("WIREFOLD_EDGE_BEAD_TRACE") == "1"

const beadChanBufferSize = 4096

type deliveredBead struct {
	val         int
	deliverTick int64
}

type BeadRun struct {
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

	readout beadReadout
}

const maxInflightBeads = beadChanBufferSize

func NewBeadRun(dwellTicks float64) *BeadRun {
	return &BeadRun{
		dwell: dwellTicks,
		inCh:  make(chan placeRequest, beadChanBufferSize),
		outCh: make(chan deliveredBead, beadChanBufferSize),

		kindToAnimClearCh: make(chan struct{}, 1),

		readout: beadReadout{breadcrumbCh: make(chan rowevent.RowEvent, 4)},
	}
}
