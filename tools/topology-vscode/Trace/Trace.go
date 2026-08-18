package Trace

import "io"

type Trace struct {
	sink    io.Writer
	onEvent func(Event)
}

func New() *Trace {
	return NewWithSink(nil)
}

func NewWithSink(sink io.Writer) *Trace {
	return NewWithSinkHook(sink, nil)
}

func NewWithSinkHook(sink io.Writer, onEvent func(Event)) *Trace {
	return &Trace{sink: sink, onEvent: onEvent}
}

func (t *Trace) SetSink(w io.Writer) {
	if t == nil {
		return
	}
	t.sink = w
}

func (t *Trace) Breadcrumb(label, node, port, value string) {
	if t == nil || t.sink == nil {
		return
	}
	b, err := marshalBreadcrumb(label, node, port, value)
	if err != nil {
		return
	}
	b = append(b, '\n')
	_, _ = t.sink.Write(b)
}
