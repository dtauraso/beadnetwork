package input

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/clock"
)

func (n *Node) runPeriodicEmit(ctx context.Context, working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) {
	emitted := 0

	lastFireTick := clk.Tick() - int64(inputCadenceTicks(n))
	for {
		if ctx.Err() != nil {
			return
		}
		now := clk.Tick()
		if (n.Repeat || emitted < len(init)) && now-lastFireTick >= int64(inputCadenceTicks(n)) {
			if n.Fire != nil {
				n.Fire()
			}
			v := popEnd(working, backup, init)
			emitBeads()
			if !n.broadcastPlace(v, now) {
				return
			}
			lastFireTick = now
			emitted++
		}

		clock.ApplySpeedNonBlocking(clk, n.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func inputCadenceTicks(n *Node) int64 {
	return cadenceTicks(n.OutCadence.Geom().Steps)
}
