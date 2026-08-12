package owners

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/clock"
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
	clock.ApplySpeedNonBlocking(c.clk, speedCh)
}

func (c *Clocks) SleepCycle(ctx context.Context) error {
	return c.clk.SleepCycle(ctx)
}

// SleepPulse waits exactly one raw pulse, unscaled by speed — the geometry
// job's cadence. It must never be mixed with SleepCycle on the same clock
// copy: one clock copy per goroutine, per the network's ownership rule.
func (c *Clocks) SleepPulse(ctx context.Context) error {
	return c.clk.SleepPulse(ctx)
}
