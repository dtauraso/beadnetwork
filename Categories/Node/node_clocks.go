package Node

import (
	"context"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
)

type Clocks struct {
	clockSrc clock.Clock

	clk clock.Clock

	humanTickMs chan int

	humanOnlyMs int
}

func NewClocks(clockSrc clock.Clock) Clocks {
	return Clocks{clockSrc: clockSrc, humanTickMs: make(chan int, 1)}
}

func (c *Clocks) Tick() int64 {
	if c.clk == nil {
		return 0
	}
	return c.clk.Tick()
}

func (c *Clocks) Use(clk clock.Clock) {
	if clk != nil {
		c.clk = clk
	}
}

func (c *Clocks) SleepCycle(ctx context.Context) error {
	return c.clk.SleepCycle(ctx)
}

func (c *Clocks) SleepPulse(ctx context.Context) error {
	return c.clk.SleepPulse(ctx)
}

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

func (g *NodeGeometry) HumanTicksOnly(tickMs int) {
	g.clocks.humanOnlyMs = max(tickMs, 1)
}

func (g *NodeGeometry) pace(ctx context.Context, clk *clock.Clock, tickMs *int) error {
	g.takeTickMs(tickMs)
	if *tickMs == 0 {
		if g.clocks.humanOnlyMs != 0 {
			panic("Node.pace: a human-tick-only geometry loop (HumanTicksOnly) reached a machine pulse; it must only ever sleep to the next human tick")
		}
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
