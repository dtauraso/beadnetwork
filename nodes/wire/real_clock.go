// real_clock.go — RealClock, the production Clock. Pacing (SleepCycle/
// SleepUntilTick) lives in sleep_cycle.go; the process-wide pulse it paces
// against lives in tick_broadcaster.go.

package wire

import "time"

// RealClock is the production Clock: SCALED wall-clock elapsed since start, floored
// to a tick via MsPerTick. "Scaled" = wall time integrated against the playback
// speed — at speed s the tick advances s× as fast as wall time. Speed changes
// accumulate the scaled elapsed up to the change instant, then continue from the
// new slope, so Tick() stays continuous and monotonic across changes (the same
// accumulate-on-transition shape the removed halt gate used, generalized from
// {0,1} to an arbitrary non-negative multiplier). Nothing waits on this clock via
// a condition variable — pacing loops call SleepCycle (wall time.After) and
// re-check Tick() themselves.
type RealClock struct {
	// RealClock is copied by VALUE: `c2 := *c1` is how a goroutine gets ITS OWN
	// independent clock, inheriting origin/accScaled/speed by value (a later SetSpeed on
	// one copy is correctly invisible to the other).
	// speed is the current playback multiplier (>= 0). Default 1.
	speed float64
	// accScaled is scaled elapsed accumulated across all PRIOR speed segments, up
	// to lastChange. The live segment (lastChange → now) is added at read time.
	accScaled time.Duration
	// lastChange is the wall instant the current speed segment began (construction
	// or the last SetSpeed).
	lastChange time.Time
	// tickCh is THIS clock's own dedicated pulse channel, subscribed from the
	// single process-wide TickBroadcaster goroutine (tick_broadcaster.go) the FIRST
	// time SleepCycle is called on this value (lazily — see sleep_cycle.go) and
	// reused on every later call by the same value. It is buffered-1 and receive-only:
	// SleepCycle blocks on it instead of on time.After, so the wall-clock wait
	// lives in exactly one goroutine (the broadcaster) for the whole process, not
	// one per clock-holder. Copy() does NOT inherit this field (each goroutine
	// that takes its own copy gets its own fresh subscription, lazily, the first
	// time IT calls SleepCycle) — see the field's zero value being fine to copy.
	tickCh <-chan struct{}
}

// NewRealClock returns a started RealClock at speed 1, anchored at the current
// monotonic instant. It does NOT subscribe to the tick broadcaster yet — that
// happens lazily, on the first SleepCycle call (see tickCh/sleep_cycle.go) — so a
// clock used only for Tick()/SetSpeed (e.g. under testing/synctest, where no
// real goroutine should be started) never touches the process-wide broadcaster.
func NewRealClock() *RealClock {
	return &RealClock{speed: 1, lastChange: time.Now()}
}

// scaledElapsed returns total scaled elapsed = accumulated prior segments + the
// live segment (wall time since lastChange × current speed). Only
// the owning goroutine ever calls this.
func (c *RealClock) scaledElapsed() time.Duration {
	live := time.Duration(float64(time.Since(c.lastChange)) * c.speed)
	total := c.accScaled + live
	if total < 0 {
		total = 0
	}
	return total
}

// Tick returns the current tick: scaled elapsed floored to ticks.
func (c *RealClock) Tick() int64 {
	return int64(c.scaledElapsed() / tickPeriod)
}

// SetSpeed sets the playback-speed multiplier. It banks the scaled elapsed of the
// segment that just ended, then starts a new segment at the new speed — so Tick()
// is continuous across the change (no jump). A negative value is clamped to 0.
func (c *RealClock) SetSpeed(speed float64) {
	if speed < 0 {
		speed = 0
	}
	now := time.Now()
	c.accScaled += time.Duration(float64(now.Sub(c.lastChange)) * c.speed)
	c.lastChange = now
	c.speed = speed
}

// Copy returns a pointer to a fresh value-copy of c: a plain struct copy (legal
// now that mu is gone — see the field comment above), inheriting the current
// speed/accScaled/lastChange by value. The caller goroutine owns the result
// from here on; nothing is shared with c or any other copy taken from it.
// tickCh is deliberately left at its zero value (nil) on the copy: the copy
// gets its own fresh broadcaster subscription lazily, the first time ITS
// SleepCycle is called (see sleep_cycle.go) — inheriting the origin's subscription
// by value would hand two goroutines the same channel, which is exactly the
// sharing this type otherwise avoids.
func (c *RealClock) Copy() Clock {
	cp := *c
	cp.tickCh = nil
	return &cp
}

// Compile-time assertion that RealClock satisfies Clock.
var _ Clock = (*RealClock)(nil)
