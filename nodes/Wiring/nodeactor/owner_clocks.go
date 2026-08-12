package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/clock"
)

func (c *nodeClocks) Tick() int64 { return c.clk.Tick() }

func (c *nodeClocks) CopyClockSrc() {
	if c.clockSrc != nil {
		c.clk = c.clockSrc.Copy()
	}
}

func (c *nodeClocks) ApplySpeed(speedCh <-chan float64) {
	clock.ApplySpeedNonBlocking(c.clk, speedCh)
}

func (c *nodeClocks) SleepCycle(ctx context.Context) error {
	return c.clk.SleepCycle(ctx)
}
