package beadanimation

type SendOutcome uint8

const (
	SendPlaced SendOutcome = iota

	SendBufferFull
)

func (bl *BeadLine) Send(v int, bp BeadPlacement, tick int64) SendOutcome {

	bl.readout.flushDroppedBreadcrumbs()
	select {
	case bl.inCh <- placeRequest{val: v, bp: bp, placementTick: tick}:
		return SendPlaced
	default:
		if bl.readout.breadcrumbCh != nil {
			select {
			case bl.readout.breadcrumbCh <- RowEvent{
				Kind: KindBreadcrumb, Label: BreadcrumbBeadPlaceBufferFull, Debug: 1,
				NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Value: int32(v),
			}:
			default:
				bl.readout.droppedBreadcrumbs++
			}
		}
		return SendBufferFull
	}
}

func (bl *BeadLine) RecvTick() (int, int64, bool) {
	select {
	case db := <-bl.outCh:
		return db.val, db.deliverTick, true
	default:
		return 0, 0, false
	}
}

func (bl *BeadLine) Recv() (int, bool) {
	v, _, ok := bl.RecvTick()
	return v, ok
}

func (bl *BeadLine) ClearInFlight() {
	if bl == nil || bl.kindToAnimClearCh == nil {
		return
	}
	select {
	case bl.kindToAnimClearCh <- struct{}{}:
	default:
	}
}

func (bl *BeadLine) applyClear() {
	select {
	case <-bl.kindToAnimClearCh:
	default:
		return
	}
	for {
		select {
		case <-bl.inCh:
		default:
			bl.inflight = nil
			return
		}
	}
}
