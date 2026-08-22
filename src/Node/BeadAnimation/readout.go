package beadanimation

import (
	"fmt"

	T "github.com/dtauraso/wirefold/src/Trace"
)

type beadReadout struct {
	StreamsActive bool

	pending []pendingBeadEvent

	breadcrumbCh chan T.RowEvent

	droppedBreadcrumbs int
}

func (bl *BeadLine) SetStreamsActive(active bool) { bl.readout.StreamsActive = active }

func (r *beadReadout) flushDroppedBreadcrumbs() {
	if r.breadcrumbCh == nil || r.droppedBreadcrumbs == 0 {
		return
	}
	select {
	case r.breadcrumbCh <- T.RowEvent{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbBeadBreadcrumbsDropped, Debug: 1,
		NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(r.droppedBreadcrumbs),
	}:
		r.droppedBreadcrumbs = 0
	default:
	}
}

func (r *beadReadout) drainBreadcrumbEvents() []T.RowEvent {
	if r.breadcrumbCh == nil {
		return nil
	}
	var out []T.RowEvent
	for {
		select {
		case ev := <-r.breadcrumbCh:
			out = append(out, ev)
		default:
			return out
		}
	}
}

func (bl *BeadLine) DrainBreadcrumbEvents() []T.RowEvent {
	return bl.readout.drainBreadcrumbEvents()
}

type pendingBeadEvent struct {
	kind       string
	value      int
	x, y, z, t float64
	gen        uint64
}

const maxPendingEvents = beadChanBufferSize

func (r *beadReadout) appendPending(ev pendingBeadEvent, owner, edge string) {
	r.pending = append(r.pending, ev)
	if len(r.pending) > maxPendingEvents {
		panic(fmt.Sprintf(
			"BeadAnimation: pending exceeded %d events on edge %q owned by node %s; the per-slot "+
				"drain (the source node's own Outs.stepBeads -> DrainPendingEvents) is not running",
			maxPendingEvents, edge, owner))
	}
}

func (r *beadReadout) drainPendingEvents() []pendingBeadEvent {
	if len(r.pending) == 0 {
		return nil
	}
	out := r.pending
	r.pending = nil
	return out
}

type PendingBeadEvent struct {
	Kind       string
	Value      int
	X, Y, Z, T float64
	Gen        uint64
}

func (bl *BeadLine) DrainPendingEvents() []PendingBeadEvent {
	internal := bl.readout.drainPendingEvents()
	if internal == nil {
		return nil
	}
	out := make([]PendingBeadEvent, len(internal))
	for i, pe := range internal {
		out[i] = PendingBeadEvent{Kind: pe.kind, Value: pe.value, X: pe.x, Y: pe.y, Z: pe.z, T: pe.t, Gen: pe.gen}
	}
	return out
}
