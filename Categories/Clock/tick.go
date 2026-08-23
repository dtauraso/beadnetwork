package clock

import "time"

const MsPerTick = 16

const tickPeriod = MsPerTick * time.Millisecond

const TickPeriod = tickPeriod

func (c *RealClock) scaledElapsed() time.Duration {
	live := time.Duration(float64(time.Since(c.lastChange)) * c.speed)
	total := c.accScaled + live
	if total < 0 {
		total = 0
	}
	return total
}

func (c *RealClock) Tick() int64 {
	return int64(c.scaledElapsed() / tickPeriod)
}
