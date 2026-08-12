package outport

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

func (o *Out) placeDrivenNoWalker(v int, tick int64) wire.SendOutcome {
	g := o.Geom()
	outcome := o.pw.Send(v, o.placementFrom(g), tick)
	if outcome != wire.SendPlaced {
		return outcome
	}
	o.flushSendEvent(v, g.Steps)
	return wire.SendPlaced
}

func (o *Out) flushSendEvent(value int, steps int) {
	if o.stream == nil {
		return
	}
	s := o.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]wire.RowEvent{{
		Kind: T.KindSend, NodeRow: s.NodeRowOf(), PortRow: o.portRow,
		TargetRow: o.targetRow, TargetPortRow: o.targetPortRow, EdgeRow: -1,
		Value:        int32(value),
		BeadSteps:    float64(steps),
		SimLatencyMs: lattice.SimLatencyMs(steps),
	}})
}

func (o *Out) Wired() bool {
	if o == nil {
		return false
	}
	return o.pw != nil
}

func NewPacedOutNoGeom(pw *wire.PacedWire, ctx context.Context, node, port string, tr *T.Trace, rule SendRule, steps int, edgeLabel string) *Out {
	return NewOutPaced(pw, ctx, node, port, tr, rule, steps, wire.WireSegment{}, edgeLabel, nil, -1, -1, -1)
}

func NewOutChanForTest(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return newOutChan(ch, node, port, tr)
}

func NewOutChanDeadEnd(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return newOutChan(ch, node, port, tr)
}

func newOutChan(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return &Out{ch: ch, node: node, port: port, trace: tr}
}

func NewOutPaced(pw *wire.PacedWire, ctx context.Context, node, port string, tr *T.Trace, rule SendRule, steps int, seg wire.WireSegment, edgeLabel string, stream func() wire.EventSink, portRow, targetRow, targetPortRow int32) *Out {
	if rule == "" {
		rule = RuleConsumeGated
	}

	fileGeom := outGeom{Steps: steps, Start: seg.Start, End: seg.End}
	o := &Out{
		pw: pw, ctx: ctx, node: node, port: port, trace: tr, Rule: rule, EdgeLabel: edgeLabel,
		geomSendSteps: make(chan int, 1),
		geomSendSeg:   make(chan wire.WireSegment, 1),
		sendCur:       fileGeom,
		stream:        stream, portRow: portRow, targetRow: targetRow, targetPortRow: targetPortRow,
	}
	return o
}
