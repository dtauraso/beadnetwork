package pulse

import (
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
)

func (a BuildArgs) Fire() func() {
	getEmitter := a.getEmitter
	return func() {
		if e := getEmitter(); e != nil {
			e.WriteEvents([]interior.RowEvent{{
				Kind: interior.KindFire, NodeRow: e.NodeRowOf(),
				PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
			}})
		}
	}
}

func (a BuildArgs) EmitNodeBeads() func(working, backup []int) {
	name, getEmitter := a.Name, a.getEmitter
	return func(working, backup []int) { interior.EmitNodeBeads(name, working, backup, getEmitter()) }
}

func (a BuildArgs) EmitHeldBead() func(held int) {
	name, getEmitter := a.Name, a.getEmitter
	return func(held int) { interior.EmitHeldBead(name, held, getEmitter()) }
}

func (a BuildArgs) EmitInputBeads() func(left, right int) {
	name, getEmitter := a.Name, a.getEmitter
	return func(left, right int) { interior.EmitInputBeads(name, left, right, getEmitter()) }
}

func (a BuildArgs) EmitRefillSlide() func(clk clock.Clock, speedCh <-chan float64, beads []int) {
	ctx, name := a.Ctx, a.Name
	return func(clk clock.Clock, speedCh <-chan float64, beads []int) {
		interior.EmitRefillSlide(ctx, name, clk, speedCh, beads)
	}
}
