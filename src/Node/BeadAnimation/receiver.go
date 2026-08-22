package beadanimation

import (
	"context"

	T "github.com/dtauraso/wirefold/src/Trace"
)

type Receiver struct {
	ch <-chan int

	line *BeadLine
	ctx  context.Context

	node string
	port string

	stream func() T.EventSink

	portRow int32
}

func (i *Receiver) PollRecv() (int, bool) {
	if i == nil {
		return 0, false
	}
	if i.line != nil {
		n, ok := i.line.Recv()
		if !ok {
			return 0, false
		}
		i.flushRecvEvent(n)
		return n, true
	}
	if i.ch == nil {
		return 0, false
	}
	select {
	case v := <-i.ch:
		i.flushRecvEvent(v)
		return v, true
	default:
		return 0, false
	}
}

func (i *Receiver) flushRecvEvent(value int) {
	if i.stream == nil {
		return
	}
	s := i.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]T.RowEvent{{
		Kind: T.KindRecv, NodeRow: s.NodeRowOf(), PortRow: i.portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: int32(value),
	}})
}

func NewInChan(ch <-chan int, node, port string, stream func() T.EventSink) *Receiver {
	return &Receiver{ch: ch, node: node, port: port, portRow: -1, stream: stream}
}

func NewInPaced(line *BeadLine, ctx context.Context, node, port string, stream func() T.EventSink, portRow int32) *Receiver {
	return &Receiver{line: line, ctx: ctx, node: node, port: port, stream: stream, portRow: portRow}
}

func (i *Receiver) HasRun() bool {
	if i == nil {
		return false
	}
	return i.line != nil
}

func (i *Receiver) Breadcrumb(event, detail string) {
	if i == nil || i.stream == nil {
		return
	}
	label, ok := breadcrumbLabelFor(event)
	if !ok {
		return
	}
	s := i.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]T.RowEvent{{
		Kind: T.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: s.NodeRowOf(), PortRow: i.portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
	}})
}

func breadcrumbLabelFor(event string) (uint8, bool) {
	switch event {
	case "topology-loaded":
		return T.BreadcrumbTopologyLoaded, true
	case "row-seed-count-mismatch":
		return T.BreadcrumbRowSeedCountMismatch, true
	case "pole-toggle-go":
		return T.BreadcrumbPoleToggleGo, true
	case "window_clear":
		return T.BreadcrumbWindowClear, true
	case "window_open":
		return T.BreadcrumbWindowOpen, true
	case "dwell_start":
		return T.BreadcrumbDwellStart, true
	case "bead-place-buffer-full":
		return T.BreadcrumbBeadPlaceBufferFull, true
	case "drag.commit":
		return T.BreadcrumbDragCommit, true
	default:
		return 0, false
	}
}
