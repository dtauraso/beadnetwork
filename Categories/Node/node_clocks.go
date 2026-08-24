package Node

import (
	"context"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
)

type Clocks struct {
	clockSrc clock.Clock

	clk clock.Clock

	speedCh <-chan float64
}

func NewClocks(clockSrc clock.Clock, clk clock.Clock) Clocks {
	return Clocks{clockSrc: clockSrc, clk: clk}
}

func (c *Clocks) Tick() int64 { return c.clk.Tick() }

func (c *Clocks) CopyClockSrc() {
	if c.clockSrc != nil {
		c.clk = c.clockSrc.Copy()
		c.clk.SpeedFrom(c.speedCh)
	}
}

func (c *Clocks) SpeedFrom(speedCh <-chan float64) {
	c.speedCh = speedCh
	if c.clk != nil {
		c.clk.SpeedFrom(speedCh)
	}
}

func (c *Clocks) SleepCycle(ctx context.Context) error {
	return c.clk.SleepCycle(ctx)
}

func (c *Clocks) SleepPulse(ctx context.Context) error {
	return c.clk.SleepPulse(ctx)
}
