package beadanimation

import (
	"context"
)

type Receiver struct {
	ch <-chan int

	line *BeadLine
	ctx  context.Context

	node string
	port string

	stream func() EventSink

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
	s.Recv(i.portRow, int32(value))
}

func NewInChan(ch <-chan int, node, port string, stream func() EventSink) *Receiver {
	return &Receiver{ch: ch, node: node, port: port, portRow: -1, stream: stream}
}

func NewInPaced(line *BeadLine, ctx context.Context, node, port string, stream func() EventSink, portRow int32) *Receiver {
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
	s := i.stream()
	if s == nil {
		return
	}
	s.Breadcrumb(event, i.portRow)
}
