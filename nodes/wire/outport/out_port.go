package outport

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
	"github.com/dtauraso/wirefold/nodes/wire"
)

type outGeom struct {
	Steps      int
	Start, End spatial.Vec3
}

type Out struct {
	ch chan<- int

	pw  *wire.PacedWire
	ctx context.Context

	node  string
	port  string
	trace *T.Trace

	sendCur outGeom

	EdgeLabel string

	Rule SendRule

	stream func() rowevent.EventSink

	portRow, targetRow, targetPortRow int32
}

func (o *Out) Geom() outGeom {
	if o == nil {
		return outGeom{}
	}
	return o.sendCur
}

func (o *Out) SetGeom(steps int, start, end spatial.Vec3) {
	if o == nil {
		return
	}
	o.sendCur = outGeom{Steps: steps, Start: start, End: end}
}

func (o *Out) placement() wire.BeadPlacement {
	return o.placementFrom(o.Geom())
}

func (o *Out) CurrentPlacement() (steps int, start, end spatial.Vec3) {
	bp := o.placement()
	return bp.Steps, bp.Start, bp.End
}

func (o *Out) placementFrom(g outGeom) wire.BeadPlacement {
	return wire.BeadPlacement{
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
