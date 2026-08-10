package gatecommon

import (
	"context"
	"time"

	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// tickDuration converts a tick count to the equivalent wall-clock time.Duration
// using the same MsPerTick conversion as defaultTick/defaultSleep below, so both
// wall-clock fallbacks agree on what a "tick" means.
func tickDuration(ticks int64) time.Duration {
	return time.Duration(ticks) * clock.MsPerTick * time.Millisecond
}

// defaultTick returns a wall-clock-derived tick function for use when GateNode.Tick
// is unset (unit tests with no loader).
func defaultTick() func() int64 {
	start := time.Now()
	return func() int64 { return int64(time.Since(start) / tickDuration(1)) }
}

// defaultSleep returns a wall-clock sleep function for use when the gate's
// output has no shared clock (unit tests with no loader): one PollIntervalTicks
// worth of wall-clock time (== one tick, since PollIntervalTicks == 1), ctx-aware.
// It receives from a dedicated tick-pulse channel subscribed ONCE here (against
// the process's one TickBroadcaster, nodes/wire/tick_broadcaster.go) rather than blocking
// on time.After — no goroutine outside the broadcaster waits on wall time.
func defaultSleep() func(ctx context.Context) error {
	tickCh := clock.NewTickChan()
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tickCh:
			return nil
		}
	}
}
