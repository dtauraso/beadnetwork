package wire

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

type outGeom struct {
	Steps      int
	Start, End Vec3
}

type Out struct {
	ch chan<- int

	pw  *PacedWire
	ctx context.Context

	node  string
	port  string
	trace *T.Trace

	geomSendSteps chan int
	geomSendSeg   chan WireSegment
	sendCur       outGeom

	EdgeLabel string

	Rule SendRule

	stream func() EventSink

	portRow, targetRow, targetPortRow int32
}

func (o *Out) Geom() outGeom {
	if o == nil {
		return outGeom{}
	}
	drainStepsNonBlocking(o.geomSendSteps, &o.sendCur.Steps)
	drainSegNonBlocking(o.geomSendSeg, &o.sendCur.Start, &o.sendCur.End)
	return o.sendCur
}

func (o *Out) publishSteps(steps int) {
	sendIntNonBlocking(o.geomSendSteps, steps)
}

func (o *Out) PublishSteps(steps int) {
	o.publishSteps(steps)
}

func (o *Out) publishSegment(start, end Vec3) {
	sendSegNonBlocking(o.geomSendSeg, WireSegment{Start: start, End: end})
}

func (o *Out) PublishSegment(start, end Vec3) {
	o.publishSegment(start, end)
}

func (o *Out) placement() beadPlacement {
	return o.placementFrom(o.Geom())
}

func (o *Out) CurrentPlacement() (steps int, start, end Vec3) {
	bp := o.placement()
	return bp.Steps, bp.Start, bp.End
}

func (o *Out) placementFrom(g outGeom) beadPlacement {
	return beadPlacement{
		Steps: g.Steps,
		Start: g.Start,
		End:   g.End,
		Node:  o.node,
		Port:  o.port,
	}
}

func (o *Out) Paced() bool {
	return o != nil && o.pw != nil
}

func (o *Out) Gated() bool {
	if o == nil {
		return true
	}
	return o.Rule != RuleFireAndForget
}

func (o *Out) placeDrivenNoWalker(v int, tick int64) SendOutcome {
	g := o.Geom()
	outcome := o.pw.Send(v, o.placementFrom(g), tick)
	if outcome != SendPlaced {
		return outcome
	}
	o.flushSendEvent(v, g.Steps)
	return SendPlaced
}

func (o *Out) flushSendEvent(value int, steps int) {
	if o.stream == nil {
		return
	}
	s := o.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]RowEvent{{
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

func NewPacedOutNoGeom(pw *PacedWire, ctx context.Context, node, port string, tr *T.Trace, rule SendRule, steps int, edgeLabel string) *Out {
	return NewOutPaced(pw, ctx, node, port, tr, rule, steps, WireSegment{}, edgeLabel, nil, -1, -1, -1)
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

func NewOutPaced(pw *PacedWire, ctx context.Context, node, port string, tr *T.Trace, rule SendRule, steps int, seg WireSegment, edgeLabel string, stream func() EventSink, portRow, targetRow, targetPortRow int32) *Out {
	if rule == "" {
		rule = RuleConsumeGated
	}

	fileGeom := outGeom{Steps: steps, Start: seg.Start, End: seg.End}
	o := &Out{
		pw: pw, ctx: ctx, node: node, port: port, trace: tr, Rule: rule, EdgeLabel: edgeLabel,
		geomSendSteps: make(chan int, 1),
		geomSendSeg:   make(chan WireSegment, 1),
		sendCur:       fileGeom,
		stream:        stream, portRow: portRow, targetRow: targetRow, targetPortRow: targetPortRow,
	}
	return o
}
