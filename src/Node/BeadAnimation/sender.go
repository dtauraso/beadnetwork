package beadanimation

import (
	T "github.com/dtauraso/wirefold/src/Trace"
	"context"

	"github.com/dtauraso/wirefold/src/spatial"
)

type outGeom struct {
	Steps      int
	SlotR      float64
	Start, End spatial.Vec3
}

type Sender struct {
	ch chan<- int

	line  *BeadLine
	ctx context.Context

	node string
	port string

	sendCur outGeom

	postedGeom chan outGeom

	EdgeLabel string

	Rule SendRule

	stream func() T.EventSink

	portRow, targetRow, targetPortRow int32
}

func (o *Sender) Geom() outGeom {
	if o == nil {
		return outGeom{}
	}
	o.applyPostedGeom()
	return o.sendCur
}

func (o *Sender) PostGeom(steps int, slotR float64, start, end spatial.Vec3) {
	if o == nil || o.postedGeom == nil {
		return
	}
	select {
	case <-o.postedGeom:
	default:
	}
	select {
	case o.postedGeom <- outGeom{Steps: steps, SlotR: slotR, Start: start, End: end}:
	default:
	}
}

func (o *Sender) applyPostedGeom() {
	if o.postedGeom == nil {
		return
	}
	select {
	case g := <-o.postedGeom:
		o.sendCur = g
	default:
	}
}

func (o *Sender) placementFrom(g outGeom) BeadPlacement {
	return BeadPlacement{
		Steps: g.Steps,
		SlotR: g.SlotR,
		Start: g.Start,
		End:   g.End,
		Node:  o.node,
		Port:  o.port,
	}
}

func (o *Sender) Paced() bool {
	return o != nil && o.line != nil
}

func (o *Sender) Gated() bool {
	if o == nil {
		return true
	}
	return o.Rule != RuleFireAndForget
}
