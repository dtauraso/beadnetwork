package wire

import "testing"

// clock_sleep_scaled_test.go — one cycle is one SCALED tick's worth of wall time, so a
// paced loop RUNS slower when the playback speed says slower.
//
// This is the property the pair scene's clock divisor could not reach before SleepCycle
// was scaled: speed applied to Tick() only, so work measured in TICKS (bead travel)
// followed the dial while work measured in CYCLES (the tilt exchange) ran at a fixed
// 62.5 Hz whatever the dial said. Asserted on pulsesPerCycle rather than by timing a real
// sleep: the arithmetic is the whole rule, and a wall-clock assertion would be a slow,
// flaky way to test a division.
func TestPulsesPerCycleScalesInverselyWithSpeed(t *testing.T) {
	cases := []struct {
		speed float64
		want  int
	}{
		// Speed 1 is the unscaled reference — one pulse, exactly what every caller got
		// before this was scaled.
		{1, 1},
		// The slider's fractional settings: half speed waits twice as long, a quarter
		// four times.
		{0.75, 2}, // 1/0.75 = 1.33, rounded UP: never shorter than the speed asks for
		{0.5, 2},
		{0.25, 4},
		// The pair scene's own effective speed at its clock divisor — the case that
		// motivated this.
		{1.0 / 16, 16},
		{1.0 / 64, 64},
		// Above 1 the broadcaster's own pulse rate is the ceiling: a faster dial shows up
		// as ticks accruing faster WITHIN a cycle, not as more cycles.
		{2, 1},
	}
	for _, c := range cases {
		clk := &RealClock{speed: c.speed}
		if got := clk.pulsesPerCycle(); got != c.want {
			t.Fatalf("speed %v: pulsesPerCycle = %d, want %d", c.speed, got, c.want)
		}
	}
}

// A stopped or extremely slow clock must still WAKE, or it can never poll its own speed
// channel to notice the dial moving back — a goroutine asleep forever cannot be restarted
// by a message it is not there to receive. Speed 0 divides to an unbounded wait, so the
// clamp is what makes zero a pause rather than a deadlock.
func TestPulsesPerCycleClampsSoAHaltedClockStillWakes(t *testing.T) {
	for _, speed := range []float64{0, -1, 1.0 / 1000} {
		clk := &RealClock{speed: speed}
		got := clk.pulsesPerCycle()
		if got != maxPulsesPerCycle {
			t.Fatalf("speed %v: pulsesPerCycle = %d, want the clamp %d", speed, got, maxPulsesPerCycle)
		}
		if got < 1 {
			t.Fatalf("speed %v: a cycle must always wait at least one pulse, got %d", speed, got)
		}
	}
}
