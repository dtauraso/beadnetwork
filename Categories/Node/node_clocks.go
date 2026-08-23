package Node

import (
	"context"
	clock "github.com/dtauraso/wirefold/Categories/Clock"

	Speed "github.com/dtauraso/wirefold/Categories/Speed"
)

type Clocks struct {
	clockSrc clock.Clock

	clk clock.Clock
}

func NewClocks(clockSrc clock.Clock, clk clock.Clock) Clocks {
	return Clocks{clockSrc: clockSrc, clk: clk}
}

func (c *Clocks) Tick() int64 { return c.clk.Tick() }

func (c *Clocks) CopyClockSrc() {
	if c.clockSrc != nil {
		c.clk = c.clockSrc.Copy()
	}
}

func (c *Clocks) ApplySpeed(speedCh <-chan float64) {
	Speed.ApplySpeedNonBlocking(c.clk, speedCh)
}

func (c *Clocks) SleepCycle(ctx context.Context) error {
	return c.clk.SleepCycle(ctx)
}

func (c *Clocks) SleepPulse(ctx context.Context) error {
	return c.clk.SleepPulse(ctx)
}
