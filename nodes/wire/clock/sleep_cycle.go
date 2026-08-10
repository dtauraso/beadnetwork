// sleep_cycle.go — RealClock's pacing methods: SleepCycle and SleepUntilTick.
// Both block on the process-wide pulse (tick_broadcaster.go), never on wall time
// directly.

package clock

import (
	"context"
	"math"
)

// SleepCycle blocks until the next tick pulse arrives on this clock's OWN
// dedicated channel (see tickCh in real_clock.go), or until ctx is done. It no
// longer waits on wall time itself — the single TickBroadcaster goroutine
// (tick_broadcaster.go) is the only thing in the process that does that; every
// SleepCycle caller just receives.
// The subscription is taken lazily, on this value's first SleepCycle call —
// exactly once per clock value, since only the value's owning goroutine ever
// calls SleepCycle on it (per-goroutine-clock.md) — so a clock that never
// paces (e.g. one only read via Tick()/SetSpeed under testing/synctest) never
// starts the process-wide broadcaster goroutine at all.
// This select carries no default case on purpose: with one it would be
// non-blocking and a caller looping around it would spin a core; without one
// the runtime parks the goroutine on both channels' wait queues at zero CPU
// until one of them has something.
// SCALED BY PLAYBACK SPEED. One cycle is one SCALED tick's worth of wall time, so the
// goroutine itself runs slower when the slider says slower — it waits pulsesPerCycle
// broadcaster pulses instead of one.
//
// This used to consume exactly one pulse regardless of speed, which made every paced loop
// in the network run at a fixed 62.5 Hz whatever the dial said. Speed scaled Tick() only,
// so anything measured in ticks (bead travel) followed the dial and anything measured in
// CYCLES did not. The pair's tilt exchange is the latter: measured on a real pair at an
// effective speed of 1/64, its whole thirteen-event run — five clicks, the kick, three
// rounds and the halt — completed at tick=0, because it finished in ~50 ms of wall time
// while one tick took about a second. Setting the scene's divisor to 16 or to 64 produced
// an identical exchange, since neither number reached the loop that was running it.
func (c *RealClock) SleepCycle(ctx context.Context) error {
	if c.tickCh == nil {
		c.tickCh = globalTickBroadcaster().Subscribe()
	}
	for i := 0; i < c.pulsesPerCycle(); i++ {
		select {
		case <-c.tickCh:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// SleepUntilTick blocks until this clock's own Tick() reaches target — see the interface
// doc. Implemented as whole SleepCycle waits rather than as one long wall wait, so it stays
// inside the ONE clock: no time.After, no deadline computed in milliseconds
// (tools/network/concurrency/check-no-wall-clock-wait.sh forbids both outside clock.go), and a playback-speed
// change during the wait is honoured on the very next cycle because the loop re-reads Tick()
// each time rather than having pre-computed when to wake.
func (c *RealClock) SleepUntilTick(ctx context.Context, target int64) error {
	for c.Tick() < target {
		if err := c.SleepCycle(ctx); err != nil {
			return err
		}
	}
	return nil
}

// maxPulsesPerCycle bounds how long one cycle can stretch — about a second at the
// broadcaster's 16 ms pulse. It is what keeps a stopped or very slow clock RESPONSIVE:
// speed 0 would otherwise divide to an unbounded wait, and a goroutine asleep forever can
// never poll its own speed channel to notice the dial moving back. So a halted network
// still wakes ~once a second, does nothing (no scaled time has passed, so nothing it
// paces advances), and stays able to hear that it should start again.
const maxPulsesPerCycle = 64

// pulsesPerCycle is how many wall pulses make one cycle at the current speed: 1/speed,
// rounded up so the wait is never shorter than the speed asks for, and clamped to
// [1, maxPulsesPerCycle]. Speed 1 gives 1 (unchanged from before this was scaled) and
// speeds above 1 also give 1 — the broadcaster's own 62.5 Hz is the ceiling on how fast a
// loop can cycle, and a faster dial shows up as ticks accruing faster within each cycle
// rather than as more cycles.
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
