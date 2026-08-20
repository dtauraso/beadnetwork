package outport

import (
	"context"

	"github.com/dtauraso/wirefold/src/spatial"
	"github.com/dtauraso/wirefold/src/Node/wire"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

type outGeom struct {
	Steps      int
	SlotR      float64
	Start, End spatial.Vec3
}

type Out struct {
	ch chan<- int

	pw  *wire.BeadRun
	ctx context.Context

	node string
	port string

	sendCur outGeom

	postedGeom chan outGeom

	EdgeLabel string

	Rule SendRule

	stream func() B.EventSink

	portRow, targetRow, targetPortRow int32
}

func (o *Out) Geom() outGeom {
	if o == nil {
		return outGeom{}
	}
	o.applyPostedGeom()
	return o.sendCur
}

func (o *Out) PostGeom(steps int, slotR float64, start, end spatial.Vec3) {
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

func (o *Out) placementFrom(g outGeom) wire.BeadPlacement {
	return wire.BeadPlacement{
		Steps: g.Steps,
		SlotR: g.SlotR,
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
