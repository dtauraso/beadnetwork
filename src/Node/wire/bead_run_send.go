package wire

import (
	"github.com/dtauraso/wirefold/src/Node/rowevent"
	T "github.com/dtauraso/wirefold/src/Trace"
)

type SendOutcome uint8

const (
	SendPlaced SendOutcome = iota

	SendBufferFull
)

func (pw *BeadRun) Send(v int, bp BeadPlacement, tick int64) SendOutcome {

	pw.readout.flushDroppedBreadcrumbs()
	select {
	case pw.inCh <- placeRequest{val: v, bp: bp, placementTick: tick}:
		return SendPlaced
	default:
		if pw.readout.Trace != nil {
			pw.readout.Trace.Breadcrumb("bead-place-buffer-full", pw.Owner, pw.Edge, "")
		}

		if pw.readout.breadcrumbCh != nil {
			select {
			case pw.readout.breadcrumbCh <- rowevent.RowEvent{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbBeadPlaceBufferFull, Debug: 1,
				NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Value: int32(v),
			}:
			default:
				pw.readout.droppedBreadcrumbs++
			}
		}
		return SendBufferFull
	}
}

func (pw *BeadRun) RecvTick() (int, int64, bool) {
	select {
	case db := <-pw.outCh:
		return db.val, db.deliverTick, true
	default:
		return 0, 0, false
	}
}

func (pw *BeadRun) Recv() (int, bool) {
	v, _, ok := pw.RecvTick()
	return v, ok
}

func (pw *BeadRun) ClearInFlight() {
	if pw == nil || pw.kindToAnimClearCh == nil {
		return
	}
	select {
	case pw.kindToAnimClearCh <- struct{}{}:
	default:
	}
}

func (pw *BeadRun) applyClear() {
	select {
	case <-pw.kindToAnimClearCh:
	default:
		return
	}
	for {
		select {
		case <-pw.inCh:
		default:
			pw.inflight = nil
			return
		}
	}
}
