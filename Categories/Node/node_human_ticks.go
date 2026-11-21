package Node

import (
	"context"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
)

func (g *NodeGeometry) RunOnHumanTicks(tickMs int) {
	select {
	case <-g.clocks.humanTickMs:
	default:
	}
	select {
	case g.clocks.humanTickMs <- max(tickMs, 1):
	default:
	}
}

func (g *NodeGeometry) pace(ctx context.Context, clk *clock.Clock, tickMs *int) error {
	g.takeTickMs(tickMs)
	if *tickMs == 0 {
		if *clk == nil {
			*clk = clock.NewRealClock()
		}
		return (*clk).SleepPulse(ctx)
	}
	v, err := clock.SleepToNextHumanTick(ctx, *tickMs, g.clocks.humanTickMs)
	*tickMs = v
	return err
}

func (g *NodeGeometry) takeTickMs(tickMs *int) {
	select {
	case v := <-g.clocks.humanTickMs:
		*tickMs = v
	default:
	}
}
