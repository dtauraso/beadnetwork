package beadanimation

import (
	"context"

	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

func (o *Sender) placeDrivenNoWalker(v int, tick int64) SendOutcome {
	g := o.Geom()
	outcome := o.pw.Send(v, o.placementFrom(g), tick)
	if outcome != SendPlaced {
		return outcome
	}
	o.flushSendEvent(v, g.Steps)
	return SendPlaced
}

func (o *Sender) flushSendEvent(value int, steps int) {
	if o.stream == nil {
		return
	}
	s := o.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]B.RowEvent{{
		Kind: B.KindSend, NodeRow: s.NodeRowOf(), PortRow: o.portRow,
		TargetRow: o.targetRow, TargetPortRow: o.targetPortRow, EdgeRow: -1,
		Value:     int32(value),
		BeadSteps: float64(steps),
	}})
}

func (o *Sender) HasRun() bool {
	if o == nil {
		return false
	}
	return o.pw != nil
}

func NewOutChanDeadEnd(ch chan<- int, node, port string) *Sender {
	return newOutChan(ch, node, port)
}

func newOutChan(ch chan<- int, node, port string) *Sender {
	return &Sender{ch: ch, node: node, port: port, postedGeom: make(chan outGeom, 1)}
}

func NewOutPaced(pw *BeadLine, ctx context.Context, node, port string, rule SendRule, edgeLabel string, stream func() B.EventSink, portRow, targetRow, targetPortRow int32) *Sender {
	if rule == "" {
		rule = RuleConsumeGated
	}

	o := &Sender{
		pw: pw, ctx: ctx, node: node, port: port, Rule: rule, EdgeLabel: edgeLabel,
		postedGeom: make(chan outGeom, 1),
		stream:     stream, portRow: portRow, targetRow: targetRow, targetPortRow: targetPortRow,
	}
	return o
}
