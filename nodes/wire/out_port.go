package wire

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
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

