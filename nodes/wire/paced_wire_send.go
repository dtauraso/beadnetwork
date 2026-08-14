package wire

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

type SendOutcome uint8

const (
	SendPlaced SendOutcome = iota

	SendBufferFull
)

func (pw *PacedWire) Send(v int, bp BeadPlacement, tick int64) SendOutcome {

	pw.readout.flushDroppedBreadcrumbs()
	select {
	case pw.inCh <- placeRequest{val: v, bp: bp, placementTick: tick}:
		return SendPlaced
	default:
		if pw.readout.Trace != nil {
			pw.readout.Trace.Breadcrumb("wire-send-buffer-full", pw.Owner, pw.Edge, "")
		}

		if pw.readout.breadcrumbCh != nil {
			select {
			case pw.readout.breadcrumbCh <- rowevent.RowEvent{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbWireSendBufferFull, Debug: 1,
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

func (pw *PacedWire) RecvTick() (int, int64, bool) {
	select {
	case db := <-pw.outCh:
		return db.val, db.deliverTick, true
	default:
		return 0, 0, false
	}
}

func (pw *PacedWire) Recv() (int, bool) {
	v, _, ok := pw.RecvTick()
	return v, ok
}

func (pw *PacedWire) ClearInFlight() {
	for {
		select {
		case <-pw.inCh:
		default:
			pw.inflight = nil
			return
		}
	}
}
