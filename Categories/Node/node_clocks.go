package Node

import (
	"context"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
)

type Clocks struct {
	clockSrc clock.Clock

	clk clock.Clock
}

func NewClocks(clockSrc clock.Clock) Clocks {
	return Clocks{clockSrc: clockSrc}
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
