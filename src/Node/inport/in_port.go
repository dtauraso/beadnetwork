package inport

import (
	"context"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

type In struct {
	ch <-chan int

	pw  *beadanimation.BeadLine
	ctx context.Context

	node string
	port string

	stream func() B.EventSink

	portRow int32
}

func (i *In) PollRecv() (int, bool) {
	if i == nil {
		return 0, false
	}
	if i.pw != nil {
		n, ok := i.pw.Recv()
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

func (i *In) flushRecvEvent(value int) {
	if i.stream == nil {
		return
	}
	s := i.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]B.RowEvent{{
		Kind: B.KindRecv, NodeRow: s.NodeRowOf(), PortRow: i.portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: int32(value),
	}})
}

func NewInChan(ch <-chan int, node, port string, stream func() B.EventSink) *In {
	return &In{ch: ch, node: node, port: port, portRow: -1, stream: stream}
}

func NewInPaced(pw *beadanimation.BeadLine, ctx context.Context, node, port string, stream func() B.EventSink, portRow int32) *In {
	return &In{pw: pw, ctx: ctx, node: node, port: port, stream: stream, portRow: portRow}
}

func (i *In) HasRun() bool {
	if i == nil {
		return false
	}
	return i.pw != nil
}

func (i *In) Breadcrumb(event, detail string) {
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
	s.WriteEvents([]B.RowEvent{{
		Kind: B.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: s.NodeRowOf(), PortRow: i.portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
	}})
}

func breadcrumbLabelFor(event string) (uint8, bool) {
	switch event {
	case "topology-loaded":
		return B.BreadcrumbTopologyLoaded, true
	case "row-seed-count-mismatch":
		return B.BreadcrumbRowSeedCountMismatch, true
	case "pole-toggle-go":
		return B.BreadcrumbPoleToggleGo, true
	case "window_clear":
		return B.BreadcrumbWindowClear, true
	case "window_open":
		return B.BreadcrumbWindowOpen, true
	case "dwell_start":
		return B.BreadcrumbDwellStart, true
	case "abc-drag":
		return B.BreadcrumbAbcDrag, true
	case "bead-place-buffer-full":
		return B.BreadcrumbBeadPlaceBufferFull, true
	case "drag.commit":
		return B.BreadcrumbDragCommit, true
	case "chain-aim":
		return B.BreadcrumbChainAim, true
	case "neighbor-center-recv":
		return B.BreadcrumbNeighborCenterRecv, true
	case "neighbor-setc-recv":
		return B.BreadcrumbNeighborSetCRecv, true
	case "bead-crud":
		return B.BreadcrumbBeadCrud, true
	case "drag-active-persist":
		return B.BreadcrumbDragActivePersist, true
	default:
		return 0, false
	}
}
