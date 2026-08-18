package outport

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/bead"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

type outGeom struct {
	Steps      int
	Start, End spatial.Vec3
}

type Out struct {
	ch chan<- int

	pw  *bead.BeadRun
	ctx context.Context

	node  string
	port  string
	trace *T.Trace

	sendCur outGeom

	postedGeom chan outGeom

	EdgeLabel string

	Rule SendRule

	stream func() rowevent.EventSink

	portRow, targetRow, targetPortRow int32
}

func (o *Out) Geom() outGeom {
	if o == nil {
		return outGeom{}
	}
	o.applyPostedGeom()
	return o.sendCur
}

func (o *Out) PostGeom(steps int, start, end spatial.Vec3) {
	if o == nil || o.postedGeom == nil {
		return
	}
	select {
	case <-o.postedGeom:
	default:
	}
	select {
	case o.postedGeom <- outGeom{Steps: steps, Start: start, End: end}:
	default:
	}
}

func (o *Out) applyPostedGeom() {
	if o.postedGeom == nil {
		return
	}
	select {
	case g := <-o.postedGeom:
		o.sendCur = g
	default:
	}
}

func (o *Out) placement() bead.BeadPlacement {
	return o.placementFrom(o.Geom())
}

func (o *Out) CurrentPlacement() (steps int, start, end spatial.Vec3) {
	bp := o.placement()
	return bp.Steps, bp.Start, bp.End
}

func (o *Out) placementFrom(g outGeom) bead.BeadPlacement {
	return bead.BeadPlacement{
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
