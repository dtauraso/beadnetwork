package kindapi

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

func (a BuildArgs) Fire() func() {
	getStream := a.getStream
	return func() {
		if s := getStream(); s != nil {
			s.WriteEvents([]rowevent.RowEvent{{
				Kind: T.KindFire, NodeRow: s.NodeRowOf(),
				PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
			}})
		}
	}
}

func (a BuildArgs) EmitNodeBeads() func(working, backup []int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(working, backup []int) { interior.EmitNodeBeads(tr, name, working, backup, getStream()) }
}

func (a BuildArgs) EmitHeldBead() func(held int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(held int) { interior.EmitHeldBead(tr, name, held, getStream()) }
}

func (a BuildArgs) EmitInputBeads() func(left, right int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(left, right int) { interior.EmitInputBeads(tr, name, left, right, getStream()) }
}

func (a BuildArgs) EmitRefillSlide() func(clk clock.Clock, speedCh <-chan float64, beads []int) {
	ctx, tr, name := a.ctx, a.tr, a.name
	return func(clk clock.Clock, speedCh <-chan float64, beads []int) {
		interior.EmitRefillSlide(ctx, tr, name, clk, speedCh, beads)
	}
}
