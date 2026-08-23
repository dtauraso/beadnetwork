package clock

import "time"

type RealClock struct {
	speed float64

	accScaled time.Duration

	lastChange time.Time

	ticker *time.Ticker

	wake <-chan struct{}
}

func NewRealClock() *RealClock {
	return &RealClock{speed: 1, lastChange: time.Now()}
}

func (c *RealClock) WakeOn(wake <-chan struct{}) {
	c.wake = wake
}

func (c *RealClock) Copy() Clock {
	cp := *c
	cp.ticker = nil
	cp.wake = nil
	return &cp
}

var _ Clock = (*RealClock)(nil)
