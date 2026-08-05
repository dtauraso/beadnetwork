package wire

import "testing"

// TestMsPerTickDivisibleByFour pins the invariant the SpeedSlider's six-value table (0,
// 0.25, 0.5, 0.75, 1, 2) depends on: MsPerTick % 4 == 0. Quarter-speed multipliers don't
// need MsPerTick itself to change (speed scales elapsed wall-nanoseconds in
// RealClock.SetSpeed/scaledElapsed, it is never a divisor of MsPerTick), but a future
// change to this constant that breaks divisibility-by-four should fail loudly here rather
// than surface as a silently-off quarter-speed step.
func TestMsPerTickDivisibleByFour(t *testing.T) {
	if MsPerTick%4 != 0 {
		t.Fatalf("MsPerTick=%d is not divisible by 4 — the speed slider's quarter steps (0.25/0.5/0.75) assume this", MsPerTick)
	}
}
