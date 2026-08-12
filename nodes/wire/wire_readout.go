package wire

import (
	"fmt"

	T "github.com/dtauraso/wirefold/Trace"
)

type wireReadout struct {
	Trace *T.Trace

	StreamsActive bool

	pending []pendingWireEvent

	breadcrumbCh chan RowEvent

	droppedBreadcrumbs int
}

func (pw *PacedWire) SetTrace(tr *T.Trace) { pw.readout.Trace = tr }

func (pw *PacedWire) SetStreamsActive(active bool) { pw.readout.StreamsActive = active }

func (r *wireReadout) flushDroppedBreadcrumbs() {
	if r.breadcrumbCh == nil || r.droppedBreadcrumbs == 0 {
		return
	}
	select {
	case r.breadcrumbCh <- RowEvent{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbWireBreadcrumbsDropped, Debug: 1,
		NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(r.droppedBreadcrumbs),
	}:
		r.droppedBreadcrumbs = 0
	default:
	}
}

func (r *wireReadout) drainBreadcrumbEvents() []RowEvent {
	if r.breadcrumbCh == nil {
		return nil
	}
	var out []RowEvent
	for {
		select {
		case ev := <-r.breadcrumbCh:
			out = append(out, ev)
		default:
			return out
		}
	}
}

func (pw *PacedWire) DrainBreadcrumbEvents() []RowEvent {
	return pw.readout.drainBreadcrumbEvents()
}

type pendingWireEvent struct {
	kind       string
	value      int
	x, y, z, t float64
	gen        uint64
}

const maxPendingEvents = wireChanBufferSize

func (r *wireReadout) appendPending(ev pendingWireEvent, target, targetHandle string) {
	r.pending = append(r.pending, ev)
	if len(r.pending) > maxPendingEvents {
		panic(fmt.Sprintf(
			"paced_wire: pending exceeded %d events on wire -> %s.%s; the per-cycle drain "+
				"(edgeMover.writeStreamFrame -> DrainPendingEvents) is not running",
			maxPendingEvents, target, targetHandle))
	}
}

func (r *wireReadout) drainPendingEvents() []pendingWireEvent {
	if len(r.pending) == 0 {
		return nil
	}
	out := r.pending
	r.pending = nil
	return out
}

type PendingWireEvent struct {
	Kind       string
	Value      int
	X, Y, Z, T float64
	Gen        uint64
}

func (pw *PacedWire) DrainPendingEvents() []PendingWireEvent {
	internal := pw.readout.drainPendingEvents()
	if internal == nil {
		return nil
	}
	out := make([]PendingWireEvent, len(internal))
	for i, pe := range internal {
		out[i] = PendingWireEvent{Kind: pe.kind, Value: pe.value, X: pe.x, Y: pe.y, Z: pe.z, T: pe.t, Gen: pe.gen}
	}
	return out
}
