package outport

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/bead"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

func (o *Out) placeDrivenNoWalker(v int, tick int64) bead.SendOutcome {
	g := o.Geom()
	outcome := o.pw.Send(v, o.placementFrom(g), tick)
	if outcome != bead.SendPlaced {
		return outcome
	}
	o.flushSendEvent(v, g.Steps)
	return bead.SendPlaced
}

func (o *Out) flushSendEvent(value int, steps int) {
	if o.stream == nil {
		return
	}
	s := o.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]rowevent.RowEvent{{
		Kind: T.KindSend, NodeRow: s.NodeRowOf(), PortRow: o.portRow,
		TargetRow: o.targetRow, TargetPortRow: o.targetPortRow, EdgeRow: -1,
		Value:     int32(value),
		BeadSteps: float64(steps),
	}})
}

func (o *Out) HasRun() bool {
	if o == nil {
		return false
	}
	return o.pw != nil
}

func NewPacedOutNoGeom(pw *bead.BeadRun, ctx context.Context, node, port string, tr *T.Trace, rule SendRule, steps int, edgeLabel string) *Out {
	return NewOutPaced(pw, ctx, node, port, tr, rule, edgeLabel, nil, -1, -1, -1)
}

func NewOutChanForTest(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return newOutChan(ch, node, port, tr)
}

func NewOutChanDeadEnd(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return newOutChan(ch, node, port, tr)
}

func newOutChan(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return &Out{ch: ch, node: node, port: port, trace: tr, postedGeom: make(chan outGeom, 1)}
}

func NewOutPaced(pw *bead.BeadRun, ctx context.Context, node, port string, tr *T.Trace, rule SendRule, edgeLabel string, stream func() rowevent.EventSink, portRow, targetRow, targetPortRow int32) *Out {
	if rule == "" {
		rule = RuleConsumeGated
	}

	o := &Out{
		pw: pw, ctx: ctx, node: node, port: port, trace: tr, Rule: rule, EdgeLabel: edgeLabel,
		postedGeom: make(chan outGeom, 1),
		stream:     stream, portRow: portRow, targetRow: targetRow, targetPortRow: targetPortRow,
	}
	return o
}
