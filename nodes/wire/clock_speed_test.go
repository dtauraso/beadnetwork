package wire

import (
	"testing"
	"testing/synctest"
	"time"
)

// clock_speed_test.go — behavioral coverage for the playback-speed scalar on RealClock.
// These pin the model the slider drives: SetSpeed multiplies tick advance, speed 0 freezes
// the tick, speed 2 doubles it, and Tick() stays continuous (never jumps back) across a
// speed change.
//
// All three run inside a synctest bubble, so `time` is the fake clock and every sleep
// advances it by exactly the requested duration. That turns the scaled advance into an
// exact integer: sleeping N tick periods at speed s advances the tick by exactly N×s.
// The wall-clock versions of these tests had to allow slack for scheduler jitter — the
// double-speed case accepted "at least 1.5×" for what the model says is exactly 2×, and
// the no-catch-up case accepted a ±5-tick band. Both are equalities now.

// TestClockSpeedFreezeAtZero: at speed 0 the tick does not advance, and it resumes cleanly
// when speed returns to 1 (no wall-clock catch-up for the frozen interval).
func TestClockSpeedFreezeAtZero(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewRealClock()
		time.Sleep(2 * tickPeriod)
		before := c.Tick()

		c.SetSpeed(0)
		frozen := c.Tick()
		time.Sleep(4 * tickPeriod)
		after := c.Tick()
		if after != frozen {
			t.Fatalf("speed 0 did not freeze the tick: frozen=%d after 4 periods=%d", frozen, after)
		}
		if after < before {
			t.Fatalf("tick went backward under freeze: before=%d after=%d", before, after)
		}

		// Resume at 1× and confirm the tick advances again from where it froze.
		c.SetSpeed(1)
		time.Sleep(2 * tickPeriod)
		resumed := c.Tick()
		if resumed <= after {
			t.Fatalf("tick did not resume after speed 1: frozen=%d resumed=%d", after, resumed)
		}
		// The frozen interval must NOT have been credited. 2 periods before the freeze
		// plus 2 after resuming = exactly 4; crediting the 4 frozen periods would give 8.
		if resumed != 4 {
			t.Fatalf("resumed tick = %d, want exactly 4 (frozen interval must not be credited; catch-up would give 8)", resumed)
		}
	})
}

// TestClockSpeedDoubleAdvancesFaster: over the same wall interval, speed 2 advances the
// tick exactly twice as far as speed 1.
func TestClockSpeedDoubleAdvancesFaster(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const periods = 8

		c1 := NewRealClock() // speed 1 (default)
		c2 := NewRealClock()
		c2.SetSpeed(2)

		time.Sleep(periods * tickPeriod)

		d1 := c1.Tick()
		d2 := c2.Tick()
		if d1 != periods {
			t.Fatalf("speed 1 advance = %d, want exactly %d", d1, periods)
		}
		if d2 != 2*periods {
			t.Fatalf("speed 2 advance = %d, want exactly %d (2x speed 1's %d)", d2, 2*periods, d1)
		}
	})
}

// TestClockSpeedContinuousAcrossChange: Tick() is monotonic non-decreasing across a
// mid-run speed change (no backward jump), which is what keeps a bead's fractional
// progress continuous when the slider moves.
func TestClockSpeedContinuousAcrossChange(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewRealClock()
		time.Sleep(3 * tickPeriod)
		a := c.Tick()
		c.SetSpeed(2)
		b := c.Tick()
		if b < a {
			t.Fatalf("tick jumped backward across a speed change: before=%d after=%d", a, b)
		}
		// A speed change banks the elapsed so far and starts a new slope — it must not
		// move the tick at the instant of the change.
		if b != a {
			t.Fatalf("speed change moved the tick at the change instant: before=%d after=%d", a, b)
		}
		c.SetSpeed(0)
		time.Sleep(2 * tickPeriod)
		c.SetSpeed(1)
		d := c.Tick()
		if d < b {
			t.Fatalf("tick jumped backward across freeze+resume: %d then %d", b, d)
		}
		// The 2 frozen periods contribute nothing, so the tick is unchanged.
		if d != b {
			t.Fatalf("frozen interval advanced the tick: before freeze=%d after resume=%d", b, d)
		}
	})
}
