package clock

import (
	"context"
	"math"
)

func (c *RealClock) SleepCycle(ctx context.Context) error {
	return c.sleepPulses(ctx, c.pulsesPerCycle())
}

func (c *RealClock) SleepPulse(ctx context.Context) error {
	return c.sleepPulses(ctx, 1)
}

func (c *RealClock) sleepPulses(ctx context.Context, n int) error {
	if c.tickCh == nil {
		c.tickCh = globalTickBroadcaster().Subscribe()
	}
	for range n {
		select {
		case <-c.tickCh:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (c *RealClock) SleepUntilTick(ctx context.Context, target int64) error {
	for c.Tick() < target {
		if err := c.SleepCycle(ctx); err != nil {
			return err
		}
	}
	return nil
}

const maxPulsesPerCycle = 64

func (c *RealClock) pulsesPerCycle() int {
	if c.speed <= 0 {
		return maxPulsesPerCycle
	}
	n := int(math.Ceil(1 / c.speed))
	if n < 1 {
		return 1
	}
	if n > maxPulsesPerCycle {
		return maxPulsesPerCycle
	}
	return n
}
