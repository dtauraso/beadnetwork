package gatecommon

import (
	"context"
	"time"

	"github.com/dtauraso/wirefold/nodes/clock"
)

func tickDuration(ticks int64) time.Duration {
	return time.Duration(ticks) * clock.MsPerTick * time.Millisecond
}

func defaultTick() func() int64 {
	start := time.Now()
	return func() int64 { return int64(time.Since(start) / tickDuration(1)) }
}

func defaultSleep() func(ctx context.Context) error {
	ticker := time.NewTicker(clock.TickPeriod)
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			return nil
		}
	}
}
