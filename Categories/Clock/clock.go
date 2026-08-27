package clock

import (
	"context"
	"math"
	"time"
)

type Clock interface {
	Tick() int64

	SleepCycle(ctx context.Context) error

	SleepPulse(ctx context.Context) error

	WakeOn(wake <-chan struct{})

	SpeedFrom(speed <-chan float64)

	Speed() float64

	Copy() Clock
}

type RealClock struct {
	speed float64

	accScaled time.Duration

	lastChange time.Time

	ticker *time.Ticker

	wake <-chan struct{}

	speedCh <-chan float64
}

func NewRealClock() *RealClock {
	return &RealClock{speed: 1, lastChange: time.Now()}
}

func (c *RealClock) WakeOn(wake <-chan struct{}) {
	c.wake = wake
}

func (c *RealClock) SpeedFrom(speed <-chan float64) {
	c.speedCh = speed
}

func (c *RealClock) Speed() float64 { return c.speed }

func (c *RealClock) applyPendingSpeed() {
	select {
	case sp := <-c.speedCh:
		c.SetSpeed(sp)
	default:
	}
}

func (c *RealClock) Copy() Clock {
	cp := *c
	cp.ticker = nil
	cp.wake = nil
	cp.speedCh = nil
	return &cp
}

var _ Clock = (*RealClock)(nil)

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

func (c *RealClock) SleepPulse(ctx context.Context) error {
	return c.sleepPulses(ctx, 1)
}

func (c *RealClock) SleepPulses(ctx context.Context, n int) error {
	if n < 1 {
		n = 1
	}
	return c.sleepPulses(ctx, n)
}

func (c *RealClock) sleepPulses(ctx context.Context, n int) error {
	c.applyPendingSpeed()
	if c.ticker == nil {
		c.ticker = time.NewTicker(tickPeriod)
	}
	for range n {
		select {
		case <-c.ticker.C:
		case <-c.wake:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (c *RealClock) SleepFor(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		d = time.Millisecond
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-c.wake:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func SleepForOrChange(ctx context.Context, d time.Duration, ch <-chan int64) (v int64, changed bool, err error) {
	if d <= 0 {
		select {
		case v = <-ch:
			return v, true, nil
		case <-ctx.Done():
			return 0, false, ctx.Err()
		}
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return 0, false, nil
	case v = <-ch:
		return v, true, nil
	case <-ctx.Done():
		return 0, false, ctx.Err()
	}
}

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

func (c *RealClock) SetSpeed(speed float64) {
	if speed < 0 {
		speed = 0
	}
	now := time.Now()
	c.accScaled += time.Duration(float64(now.Sub(c.lastChange)) * c.speed)
	c.lastChange = now
	c.speed = speed
}
