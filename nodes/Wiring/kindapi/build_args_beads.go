package kindapi

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
)

func (a BuildArgs) Fire() func() {
	getEmitter := a.getEmitter
	return func() {
		if e := getEmitter(); e != nil {
			e.WriteEvents([]rowevent.RowEvent{{
				Kind: T.KindFire, NodeRow: e.NodeRowOf(),
				PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
			}})
		}
	}
}

func (a BuildArgs) EmitNodeBeads() func(working, backup []int) {
	tr, name, getEmitter := a.tr, a.name, a.getEmitter
	return func(working, backup []int) { interior.EmitNodeBeads(tr, name, working, backup, getEmitter()) }
}

func (a BuildArgs) EmitHeldBead() func(held int) {
	tr, name, getEmitter := a.tr, a.name, a.getEmitter
	return func(held int) { interior.EmitHeldBead(tr, name, held, getEmitter()) }
}

func (a BuildArgs) EmitInputBeads() func(left, right int) {
	tr, name, getEmitter := a.tr, a.name, a.getEmitter
	return func(left, right int) { interior.EmitInputBeads(tr, name, left, right, getEmitter()) }
}

func (a BuildArgs) EmitRefillSlide() func(clk clock.Clock, speedCh <-chan float64, beads []int) {
	ctx, tr, name := a.ctx, a.tr, a.name
	return func(clk clock.Clock, speedCh <-chan float64, beads []int) {
		interior.EmitRefillSlide(ctx, tr, name, clk, speedCh, beads)
	}
}
