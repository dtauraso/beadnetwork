package outport

import "github.com/dtauraso/wirefold/src/Node/wire"

type DriveOutcome uint8

const (
	DrivePlaced DriveOutcome = iota

	DriveSentChan

	DriveBufferFull

	DriveFailed
)

type DriveItem struct {
	outcome DriveOutcome
}

func (di DriveItem) Live() bool {
	return di.outcome == DrivePlaced
}

func (di DriveItem) Failed() bool {
	return di.outcome == DriveFailed
}

func (di DriveItem) BufferFull() bool {
	return di.outcome == DriveBufferFull
}

func (o *Out) PlaceDrivenAt(v int, tick int64) DriveItem {
	if o == nil {
		return DriveItem{outcome: DriveFailed}
	}
	if o.pw != nil {
		switch o.placeDrivenNoWalker(v, tick) {
		case wire.SendPlaced:
			return DriveItem{outcome: DrivePlaced}
		default:
			return DriveItem{outcome: DriveBufferFull}
		}
	}

	if o.ch != nil {
		select {
		case o.ch <- v:
			o.flushSendEvent(v, 0)
		default:
		}
		return DriveItem{outcome: DriveSentChan}
	}
	return DriveItem{outcome: DriveSentChan}
}
