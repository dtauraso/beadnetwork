package Wiring

import "testing"

// tilt_edit_speed_test.go — SETTING a tilt runs the clocks at human speed; STARTING or
// RESETTING puts the slider's speed back.
//
// The two speeds have to differ for the override to be observable at all, which is the
// point: since SleepCycle became speed-scaled, a scene's divisor stretches the very loop
// that drains a ▲/▼ click, so without this a click waits a scaled cycle to be noticed.
//
// Asserted on the ARITHMETIC both call sites use (HumanEditSpeed vs md.SliderSpeed) rather
// than by driving the reader and watching channels — what a click is worth is one
// goroutine's own decision, and a delivery test would assert what the channels already
// guarantee (docs/testing-shape.md).
func TestHumanEditSpeedIsUnscaledAndDiffersFromASlowedScene(t *testing.T) {
	if HumanEditSpeed != 1 {
		t.Fatalf("HumanEditSpeed = %v, want 1 — setting an angle is an interaction, not a simulation", HumanEditSpeed)
	}
	// A scene with a divisor (the pair) must resolve to something SLOWER than the edit
	// speed, or the override would be a no-op there and the click would still wait.
	pairDivisor := SceneClockDivisor("/anywhere/topology-pair")
	slider := EffectiveClockSpeed(1, pairDivisor)
	if slider >= HumanEditSpeed {
		t.Fatalf("pair slider speed %v is not slower than HumanEditSpeed %v — the override buys nothing", slider, HumanEditSpeed)
	}
}

// SliderSpeed is what the restore sends, and it must be exactly what a live slider change
// would have sent — the same userSpeed through the same divisor. If these two ever
// diverged, a start/reset would silently leave the network at some third speed.
func TestSliderSpeedMatchesALiveSliderChange(t *testing.T) {
	for _, userSpeed := range []float64{0, 0.25, 0.5, 0.75, 1, 2} {
		for _, divisor := range []float64{1, 4, 64} {
			md := &MoveDispatch{}
			md.ui.speed = userSpeed
			md.ui.clockDivisor = divisor
			want := EffectiveClockSpeed(userSpeed, divisor)
			if got := md.SliderSpeed(); got != want {
				t.Fatalf("userSpeed=%v divisor=%v: SliderSpeed = %v, want %v", userSpeed, divisor, got, want)
			}
		}
	}
}

// BroadcastSpeed must reach EVERY sink — a clock left un-told keeps pacing at whatever it
// last heard, which is how one goroutine ends up running at a different speed from the
// rest of the network.
func TestBroadcastSpeedReachesEverySink(t *testing.T) {
	sinks := []chan float64{make(chan float64, 1), make(chan float64, 1), make(chan float64, 1)}
	BroadcastSpeed(sinks, 0.25)
	for i, ch := range sinks {
		select {
		case got := <-ch:
			if got != 0.25 {
				t.Fatalf("sink %d received %v, want 0.25", i, got)
			}
		default:
			t.Fatalf("sink %d received nothing", i)
		}
	}
}
