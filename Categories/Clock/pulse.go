package clock

import (
	"context"
	"time"
)

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
