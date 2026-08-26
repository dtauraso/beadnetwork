package NodePhi

import (
	"context"

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
	clk.SpeedFrom(n.plumb.SpeedCh)
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

		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}
