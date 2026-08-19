package input

import (
	"context"

	"github.com/dtauraso/wirefold/src/Node/clock"
)

func (n *Node) runPeriodicEmit(ctx context.Context, working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) {
	emitted := 0

	lastFireTick := clk.Tick() - int64(inputCadenceTicks(n))
	n.runStepLoop(ctx, clk, func() bool {
		now := clk.Tick()
		if (n.Repeat || emitted < len(init)) && now-lastFireTick >= int64(inputCadenceTicks(n)) {
			if n.Fire != nil {
				n.Fire()
			}
			v := popEnd(working, backup, init)
			emitBeads()
			if !n.broadcastPlace(v, now) {
				return false
			}
			lastFireTick = now
			emitted++
		}
		return true
	})
}

func inputCadenceTicks(n *Node) int64 {
	return cadenceTicks(n.OutCadence.Geom().Steps)
}
