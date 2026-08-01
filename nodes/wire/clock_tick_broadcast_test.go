package wire

import (
	"context"
	"testing"
	"time"
)

// clock_tick_broadcast_test.go — pins the two-clock-beads Phase A property: SleepCycle no
// longer waits on wall time itself (nodes/wire/clock.go's TickBroadcaster is the only
// goroutine in the process that does that); every caller blocks on RECEIVE from a
// dedicated channel instead. These tests use REAL time (not synctest) because they are
// pinning a property of the real background broadcaster goroutine.

// TestSleepCycleReceivesTick: SleepCycle actually returns once the broadcaster's next
// pulse arrives — the replacement mechanism still paces at roughly MsPerTick.
func TestSleepCycleReceivesTick(t *testing.T) {
	clk := NewRealClock()
	start := time.Now()
	if err := clk.SleepCycle(context.Background()); err != nil {
		t.Fatalf("SleepCycle returned error: %v", err)
	}
	elapsed := time.Since(start)
	// Generous bound: one real tick plus scheduling slack, well under 10x tickPeriod.
	if elapsed > 10*tickPeriod {
		t.Fatalf("SleepCycle took %v, want roughly one tickPeriod (%v)", elapsed, tickPeriod)
	}
}

// TestSleepCycleUnblocksOnContextCancel: a goroutine parked in SleepCycle waiting for its
// next tick must still unblock immediately when ctx is cancelled — it must not need to
// wait out a full tick first. This is the ctx-cancellation half of "no sleeping": SleepCycle
// is a select over the tick channel AND ctx.Done(), so cancellation is serviced without
// waiting on the tick.
func TestSleepCycleUnblocksOnContextCancel(t *testing.T) {
	clk := NewRealClock()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- clk.SleepCycle(ctx)
	}()

	// Cancel immediately, well inside the first tick period, and confirm SleepCycle
	// returns right away rather than waiting for the tick that would otherwise arrive
	// ~tickPeriod later.
	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("SleepCycle returned %v, want context.Canceled", err)
		}
		if elapsed := time.Since(start); elapsed > tickPeriod {
			t.Fatalf("SleepCycle took %v to unblock on cancel, want well under one tickPeriod (%v)", elapsed, tickPeriod)
		}
	case <-time.After(2 * tickPeriod):
		t.Fatal("SleepCycle did not unblock on ctx cancellation")
	}
}

// TestTickWaitServicesOtherChannelImmediately is the behavioural statement of "no
// sleeping" (PLAN.md): a goroutine waiting for its next tick, parked in a select that also
// carries another channel, must service that other channel immediately — it must not have
// to wait out the tick first. This is exactly what a wall-clock time.After (the old
// SleepCycle) could NOT do: nothing else could be serviced until the sleep elapsed. Because
// the tick is now a channel receive (NewTickChan/tickCh), it composes into a select
// alongside any other channel a caller cares about, with no default case (blocking, no
// spin) — the same shape a mover's own pacing point would use.
func TestTickWaitServicesOtherChannelImmediately(t *testing.T) {
	tickCh := NewTickChan()
	other := make(chan struct{})

	go func() {
		// Give the receiver a moment to be parked in the select below, then send on
		// the OTHER channel — well before the next real tick (tickPeriod away).
		time.Sleep(tickPeriod / 4)
		other <- struct{}{}
	}()

	start := time.Now()
	select {
	case <-tickCh:
		t.Fatal("tick fired before the other channel — test did not exercise the intended race")
	case <-other:
		// Serviced the other channel without waiting for the tick.
	}
	if elapsed := time.Since(start); elapsed >= tickPeriod {
		t.Fatalf("took %v to service the other channel, want well under one tickPeriod (%v) — tick wait blocked it", elapsed, tickPeriod)
	}
}
