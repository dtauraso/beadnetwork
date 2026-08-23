package clock

import (
	"context"
	"math"
)

const maxPulsesPerCycle = 64

const PausedCycleMultiple = maxPulsesPerCycle

func (c *RealClock) SleepCycle(ctx context.Context) error {
	return c.sleepPulses(ctx, c.pulsesPerCycle())
}

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
