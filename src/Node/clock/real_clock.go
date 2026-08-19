package clock

import "time"

type RealClock struct {
	speed float64

	accScaled time.Duration

	lastChange time.Time

	ticker *time.Ticker

	wake <-chan struct{}
}

func (c *RealClock) WakeOn(wake <-chan struct{}) {
	c.wake = wake
}

func NewRealClock() *RealClock {
	return &RealClock{speed: 1, lastChange: time.Now()}
}

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

func (c *RealClock) SetSpeed(speed float64) {
	if speed < 0 {
		speed = 0
	}
	now := time.Now()
	c.accScaled += time.Duration(float64(now.Sub(c.lastChange)) * c.speed)
	c.lastChange = now
	c.speed = speed
}

func (c *RealClock) Copy() Clock {
	cp := *c
	cp.ticker = nil
	cp.wake = nil
	return &cp
}

var _ Clock = (*RealClock)(nil)
