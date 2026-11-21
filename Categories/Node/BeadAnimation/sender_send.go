package beadanimation

import (
	"context"
)

func (o *Sender) placeDrivenNoWalker(v int, tick int64) SendOutcome {
	g := o.Geom()
	outcome := o.line.Send(v, o.placementFrom(g), tick)
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
	s.Send(o.portRow, o.targetRow, o.targetPortRow, int32(value), float64(steps))
}

func (o *Sender) HasRun() bool {
	if o == nil {
		return false
	}
	return o.line != nil
}

func NewOutChanDeadEnd(ch chan<- int, node, port string) *Sender {
	return newOutChan(ch, node, port)
}

func newOutChan(ch chan<- int, node, port string) *Sender {
	return &Sender{ch: ch, node: node, port: port, postedGeom: make(chan outGeom, 1)}
}

func NewOutPaced(line *BeadLine, ctx context.Context, node, port string, rule SendRule, edgeLabel string, stream func() EventSink, portRow, targetRow, targetPortRow int32) *Sender {
	if rule == "" {
		rule = RuleConsumeGated
	}

	o := &Sender{
		line: line, ctx: ctx, node: node, port: port, Rule: rule, EdgeLabel: edgeLabel,
		postedGeom: make(chan outGeom, 1),
		stream:     stream, portRow: portRow, targetRow: targetRow, targetPortRow: targetPortRow,
	}
	return o
}

type DriveOutcome uint8

const (
	DrivePlaced DriveOutcome = iota

	DriveSentChan

	DriveBufferFull

	DriveFailed
)

type DriveItem struct {
	outcome DriveOutcome
}

func (di DriveItem) Live() bool {
	return di.outcome == DrivePlaced
}

func (di DriveItem) Failed() bool {
	return di.outcome == DriveFailed
}

func (di DriveItem) BufferFull() bool {
	return di.outcome == DriveBufferFull
}

func (o *Sender) PlaceDrivenAt(v int, tick int64) DriveItem {
	if o == nil {
		return DriveItem{outcome: DriveFailed}
	}
	if o.line != nil {
		switch o.placeDrivenNoWalker(v, tick) {
		case SendPlaced:
			return DriveItem{outcome: DrivePlaced}
		default:
			return DriveItem{outcome: DriveBufferFull}
		}
	}

	if o.ch != nil {
		select {
		case o.ch <- v:
			o.flushSendEvent(v, 0)
		default:
		}
		return DriveItem{outcome: DriveSentChan}
	}
	return DriveItem{outcome: DriveSentChan}
}
