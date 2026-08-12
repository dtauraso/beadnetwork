package wire

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
)

type In struct {
	ch <-chan int

	pw  *PacedWire
	ctx context.Context

	node  string
	port  string
	trace *T.Trace

	stream func() EventSink

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
	s.WriteEvents([]RowEvent{{
		Kind: T.KindRecv, NodeRow: s.NodeRowOf(), PortRow: i.portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: int32(value),
	}})
}

func NewInChan(ch <-chan int, node, port string, tr *T.Trace, stream func() EventSink) *In {
	return &In{ch: ch, node: node, port: port, trace: tr, portRow: -1, stream: stream}
}

func NewInPaced(pw *PacedWire, ctx context.Context, node, port string, tr *T.Trace, stream func() EventSink, portRow int32) *In {
	return &In{pw: pw, ctx: ctx, node: node, port: port, trace: tr, stream: stream, portRow: portRow}
}

func (i *In) Wired() bool {
	if i == nil {
		return false
	}
	return i.pw != nil
}

func (i *In) Breadcrumb(event, detail string) {
	if i == nil || i.trace == nil {
		return
	}
	node, port := i.node, i.port
	if i.pw != nil {
		node, port = i.pw.Target, i.pw.TargetHandle
	}
	i.trace.Breadcrumb(event, node, port, detail)

	if i.stream == nil {
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
	s.WriteEvents([]RowEvent{{
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
	case "abc-drag":
		return T.BreadcrumbAbcDrag, true
	case "wire-send-buffer-full":
		return T.BreadcrumbWireSendBufferFull, true
	case "drag.commit":
		return T.BreadcrumbDragCommit, true
	case "chain-aim":
		return T.BreadcrumbChainAim, true
	case "neighbor-center-recv":
		return T.BreadcrumbNeighborCenterRecv, true
	case "neighbor-setc-recv":
		return T.BreadcrumbNeighborSetCRecv, true
	case "bead-crud":
		return T.BreadcrumbBeadCrud, true
	default:
		return 0, false
	}
}
