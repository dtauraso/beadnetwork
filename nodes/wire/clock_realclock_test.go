package wire

import (
	"testing"
	"testing/synctest"
	"time"
)

// clock_realclock_test.go — behavioral coverage for RealClock (the production Clock).
//
// Runs inside a synctest bubble, so `time` is the bubble's FAKE clock: time.Sleep
// advances it by exactly the requested duration with no scheduler jitter. RealClock is a
// pure function of time.Now()/time.Since (see clock.go), so it reads the fake clock too,
// and the tick after a sleep of N tick periods is exactly N — not "at least 1". The
// assertion below is therefore an equality, where against the wall clock it could only
// be an inequality with slack.

// TestRealClockTickMonotonic: Tick() never goes backward, and advances by EXACTLY the
// slept number of tick periods.
func TestRealClockTickMonotonic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewRealClock()
		a := c.Tick()
		time.Sleep(2 * tickPeriod)
		b := c.Tick()
		if b < a {
			t.Fatalf("Tick() went backward: first=%d second=%d", a, b)
		}
		if b != a+2 {
			t.Fatalf("Tick() advanced %d ticks across a 2-tick sleep, want exactly 2: first=%d second=%d", b-a, a, b)
		}
	})
}
