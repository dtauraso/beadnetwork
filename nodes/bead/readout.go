package bead

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/rowevent"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

type beadReadout struct {
	Trace *T.Trace

	StreamsActive bool

	pending []pendingBeadEvent

	breadcrumbCh chan rowevent.RowEvent

	droppedBreadcrumbs int
}

func (pw *BeadRun) SetTrace(tr *T.Trace) { pw.readout.Trace = tr }

func (pw *BeadRun) SetStreamsActive(active bool) { pw.readout.StreamsActive = active }

func (r *beadReadout) flushDroppedBreadcrumbs() {
	if r.breadcrumbCh == nil || r.droppedBreadcrumbs == 0 {
		return
	}
	select {
	case r.breadcrumbCh <- rowevent.RowEvent{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbBeadBreadcrumbsDropped, Debug: 1,
		NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(r.droppedBreadcrumbs),
	}:
		r.droppedBreadcrumbs = 0
	default:
	}
}

func (r *beadReadout) drainBreadcrumbEvents() []rowevent.RowEvent {
	if r.breadcrumbCh == nil {
		return nil
	}
	var out []rowevent.RowEvent
	for {
		select {
		case ev := <-r.breadcrumbCh:
			out = append(out, ev)
		default:
			return out
		}
	}
}

func (pw *BeadRun) DrainBreadcrumbEvents() []rowevent.RowEvent {
	return pw.readout.drainBreadcrumbEvents()
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
			"bead_run: pending exceeded %d events on edge %q owned by node %s; the per-slot "+
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

func (pw *BeadRun) DrainPendingEvents() []PendingBeadEvent {
	internal := pw.readout.drainPendingEvents()
	if internal == nil {
		return nil
	}
	out := make([]PendingBeadEvent, len(internal))
	for i, pe := range internal {
		out[i] = PendingBeadEvent{Kind: pe.kind, Value: pe.value, X: pe.x, Y: pe.y, Z: pe.z, T: pe.t, Gen: pe.gen}
	}
	return out
}
