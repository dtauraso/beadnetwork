// build_args_beads.go — BuildArgs methods that hand a node its own interior BEAD EMISSION
// closures (Fire and the Emit* family), all of which write to this node's own interior
// stream via a.getStream. Split out of build_args.go — see that file's header.

package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// Fire returns this node's fire-trace closure. The node name is captured, so a node
// cannot mis-name itself in the trace. The event lands on this node's OWN interior
// stream; nil-safe when the node has no dedicated interior fd (test builds).
func (a BuildArgs) Fire() func() {
	getStream := a.getStream
	return func() {
		if s := getStream(); s != nil {
			s.WriteEvents([]wire.RowEvent{{
				Kind: T.KindFire, NodeRow: s.NodeRowOf(),
				PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
			}})
		}
	}
}

// EmitNodeBeads returns the interior working/backup bead emitter.
func (a BuildArgs) EmitNodeBeads() func(working, backup []int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(working, backup []int) { interior.EmitNodeBeads(tr, name, working, backup, getStream()) }
}

// EmitHeldBead returns the single centered held-value bead emitter (held == NoValue
// renders as an empty interior).
func (a BuildArgs) EmitHeldBead() func(held int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(held int) { interior.EmitHeldBead(tr, name, held, getStream()) }
}

// EmitInputBeads returns a gate's two-sided held-input bead emitter.
func (a BuildArgs) EmitInputBeads() func(left, right int) {
	tr, name, getStream := a.tr, a.name, a.getStream
	return func(left, right int) { interior.EmitInputBeads(tr, name, left, right, getStream()) }
}

// EmitRefillSlide returns the clock-paced refill-slide emitter. The clock and speed
// channel are supplied by the CALLER at invocation time — its own already-Copy()'d clock
// and its own SpeedCh — never captured here, per per-goroutine-clock.md.
func (a BuildArgs) EmitRefillSlide() func(clk clock.Clock, speedCh <-chan float64, beads []int) {
	ctx, tr, name := a.ctx, a.tr, a.name
	return func(clk clock.Clock, speedCh <-chan float64, beads []int) {
		interior.EmitRefillSlide(ctx, tr, name, clk, speedCh, beads)
	}
}
