package PairNode

import (
	"context"

	Speed "github.com/dtauraso/wirefold/Categories/Speed"
)

type Node struct {
	plumb nodePlumbing

	tilt tiltHeld

	lattice latticeState

	vec vectorExchange

	rest restCounters
}

func (n *Node) Update(ctx context.Context) {
	n.openingEmit()

	clk := n.clock().Copy()
	n.plumb.Self.StartRule(ctx, clk)

	for {
		if ctx.Err() != nil {
			return
		}

		n.paceOnBeadArrival()
		n.drainTiltEdit(clk)
		n.drainLattice()

		n.handleVectorCycle(clk.Tick())

		n.plumb.Self.Step(ctx, clk.Tick())

		Speed.ApplySpeedNonBlocking(clk, n.plumb.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}
