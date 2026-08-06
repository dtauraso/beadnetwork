// clock.go — the single human-speed clock the network reads.
//
// MODEL.md pins this: there is exactly one clock — the system monotonic clock
// read through a fixed SCALE so it advances in integer TICKS at human-watchable
// speed (`tick = ⌊(systemNow − start) × scale⌋`). All timing in the network is
// tick counts, never wall-clock durations: goroutines pace themselves with
// SleepCycle, which blocks for exactly one clock cycle. A bead crossing an edge
// takes `ticksToCross = steps * DwellTicksPerBead` ticks (steps: the edge's own
// bead-step count, docs/bead-lattice.md "The count"; DwellTicksPerBead: the one
// uniform per-step dwell, bead_lattice.go); node
// processing windows are tick counts. There is no separate render cadence — the
// tick IS the animation clock.
//
// SleepCycle no longer blocks on a per-caller time.After: it receives from a dedicated
// channel subscribed against the process's ONE TickBroadcaster goroutine (below), which
// is the only thing that ever waits on wall time. Every pacing loop still calls
// SleepCycle exactly the same way; only what it blocks on changed.
//
// SCALE arithmetic (behavior-preserving vs. the retired wall-clock model, and
// vs. the later retired arc-length model — docs/bead-lattice.md superseded both):
// the original model sampled bead positions every 16 ms. We pick one tick ≈ one
// old 16 ms sample:
//
//	MsPerTick = 16   ⇒   scale = 1 tick / 16 ms = 62.5 ticks/sec.
//
// So a bead visits ~the same number of positions in ~the same wall time, and
// pause/resume look identical.
//
// The model is sleep-only: pacing loops call SleepCycle to wait exactly one
// clock cycle rather than blocking on a target tick. RealClock is the single
// production Clock implementation.
//
// The clock is free-running: there is no play/pause gate (that feature was removed
// end-to-end), so the tick advances monotonically with wall time for the life of
// the process.

package wire

import (
	"context"
	"math"
	"sync"
	"time"
)

// MsPerTick is the scale of the human-speed clock: one tick spans this many
// wall-milliseconds while running (scale = 1/MsPerTick ticks per ms). 16 ms/tick
// = 62.5 ticks/sec, matching the retired 16 ms position-sample cadence so visible
// bead speed is unchanged.
const MsPerTick = 16

// tickPeriod is MsPerTick as a Duration (the wall span of one running tick).
const tickPeriod = MsPerTick * time.Millisecond

// TickPeriod is tickPeriod's exported mirror, for callers in another package
// (speed_delivery_test.go in nodes/Wiring) that need the wall span of one tick.
const TickPeriod = tickPeriod

// Clock is the one human-speed clock the network reads. Tick() returns the
// current integer tick, advancing at the current playback speed.
type Clock interface {
	// Tick returns the current tick since the clock started. Ticks accrue as
	// SCALED wall time — wall elapsed integrated against the playback speed (see
	// SetSpeed) — so at speed 1 it is plain wall time, at 2 it advances twice as
	// fast, and at 0 it holds. It is monotonic non-decreasing for the process life.
	Tick() int64
	// SleepCycle blocks for exactly one WALL clock cycle (or until ctx is done).
	// It is the primitive for one-cycle pacing loops; it does not read Tick()
	// itself, so it is immune to a tick advancing between the call and the wait.
	// It sleeps WALL time regardless of playback speed — the loop re-reads Tick()
	// to see how many scaled ticks actually elapsed, so speed scaling lives in
	// Tick(), not in the sleep cadence.
	SleepCycle(ctx context.Context) error
	// Copy returns a clock a single goroutine can OWN from this point on. Per
	// per-goroutine-clock.md: a goroutine calls Copy() exactly ONCE, at its own
	// start, and uses only the returned clock thereafter — never a second call
	// mid-loop, and the returned clock must never be handed to a second
	// goroutine (that would just re-share the same object under a new name).
	// *RealClock returns a pointer to a fresh value-copy of itself, so the two
	// clocks share no memory: the copy inherits the origin/accScaled/speed by
	// value and from then on a speed change on one is invisible to the other,
	// correctly, with nothing left to lock.
	Copy() Clock
}

// SetSpeed is deliberately NOT on the Clock interface (per-goroutine-clock.md
// item 4): once a clock is a per-goroutine COPY, nothing outside the goroutine
// that owns it may mutate it — a second goroutine reaching in to call SetSpeed
// on someone else's copy is exactly the shared-object shape this model removes.
// (*RealClock).SetSpeed stays a concrete, exported method: the owning goroutine
// still needs to apply a speed change to ITS OWN copy. How a speed change is
// DELIVERED to every live copy is a separate, not-yet-built step (see
// per-goroutine-clock.md "Delivery"); stdin_reader.go's current SetSpeed call
// site type-asserts down to *RealClock and is a KNOWN, documented no-op today
// (see its own comment) — not this interface's concern.

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
	// single process-wide TickBroadcaster goroutine (see below) the FIRST time
	// SleepCycle is called on this value (lazily — see SleepCycle) and reused on
	// every later call by the same value. It is buffered-1 and receive-only:
	// SleepCycle blocks on it instead of on time.After, so the wall-clock wait
	// lives in exactly one goroutine (the broadcaster) for the whole process, not
	// one per clock-holder. Copy() does NOT inherit this field (each goroutine
	// that takes its own copy gets its own fresh subscription, lazily, the first
	// time IT calls SleepCycle) — see the field's zero value being fine to copy.
	tickCh <-chan struct{}
}

// NewRealClock returns a started RealClock at speed 1, anchored at the current
// monotonic instant. It does NOT subscribe to the tick broadcaster yet — that
// happens lazily, on the first SleepCycle call (see tickCh/SleepCycle) — so a
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
// SleepCycle is called (see SleepCycle) — inheriting the origin's subscription
// by value would hand two goroutines the same channel, which is exactly the
// sharing this type otherwise avoids.
func (c *RealClock) Copy() Clock {
	cp := *c
	cp.tickCh = nil
	return &cp
}

// SleepCycle blocks until the next tick pulse arrives on this clock's OWN
// dedicated channel (see tickCh), or until ctx is done. It no longer waits on
// wall time itself — the single TickBroadcaster goroutine (below) is the only
// thing in the process that does that; every SleepCycle caller just receives.
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

// Compile-time assertion that RealClock satisfies Clock.
var _ Clock = (*RealClock)(nil)

// TickBroadcaster is the ONE thing in the process that waits on wall time. Its
// single goroutine (run) owns a time.Ticker and, on every fire, pushes a pulse
// to every subscriber's own dedicated channel — non-blockingly, so a slow
// subscriber never stalls the broadcaster or any other subscriber. Every other
// goroutine in the network (movers, node loops, DriveHeld, gate loops) blocks
// on RECEIVE from its own subscription instead of sleeping — see clock.go's
// package doc and PLAN.md "No sleeping".
type TickBroadcaster struct {
	// register is how a new subscriber hands the broadcaster its channel to add
	// to the fan-out set. Unbuffered: the broadcaster's run loop is always
	// selecting on it (or the ticker), so a send here never blocks past one
	// broadcaster loop iteration.
	register chan chan struct{}
}

// startTickBroadcaster starts the broadcaster's one goroutine and returns a
// handle new subscribers register against. It runs for the life of the
// process — there is exactly one of these (see globalTickBroadcaster), no
// per-goroutine tickers anywhere else.
func startTickBroadcaster() *TickBroadcaster {
	tb := &TickBroadcaster{register: make(chan chan struct{})}
	go tb.run()
	return tb
}

// run is the ONE goroutine in the process that ever calls time.NewTicker (or
// blocks on it). It owns the fan-out subscriber list — nothing else reads or
// writes subs, so it needs no lock.
func (tb *TickBroadcaster) run() {
	ticker := time.NewTicker(tickPeriod)
	defer ticker.Stop()
	var subs []chan struct{}
	for {
		select {
		case ch := <-tb.register:
			subs = append(subs, ch)
		case <-ticker.C:
			for _, ch := range subs {
				// Non-blocking, coalescing send: a subscriber that has not yet
				// drained the previous pulse just sees "at least one tick has
				// fired since I last checked" rather than backing up a queue —
				// the same latest-wins shape as SendSpeedNonBlocking/
				// SendLatestNonBlocking elsewhere in this file.
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}
}

// Subscribe returns a fresh, dedicated buffered-1 channel that receives a
// pulse once per wall tick from this broadcaster. Call once per goroutine (or
// once per Clock value, which is the same thing under per-goroutine-clock.md);
// the returned channel is owned by the caller from then on and must not be
// shared with a second goroutine.
func (tb *TickBroadcaster) Subscribe() <-chan struct{} {
	pulseCh := make(chan struct{}, 1) // chan-name-ok: internal broadcaster->subscriber pulse, not a node-to-node wire
	tb.register <- pulseCh
	return pulseCh
}

var (
	tickBroadcasterOnce sync.Once
	tickBroadcasterInst *TickBroadcaster
)

// globalTickBroadcaster lazily starts (sync.Once — coordinates goroutine
// startup, not shared mutable state, so it stays inside check-no-network-locks'
// allowance for sync.Once) the process-wide single clock goroutine on first
// use and returns it thereafter. Every RealClock (NewRealClock and Copy)
// subscribes against this one instance, so however many clock-holders the
// process has, exactly one goroutine ever waits on wall time.
func globalTickBroadcaster() *TickBroadcaster {
	tickBroadcasterOnce.Do(func() {
		tickBroadcasterInst = startTickBroadcaster()
	})
	return tickBroadcasterInst
}

// NewTickChan returns a fresh dedicated tick-pulse channel from the global
// broadcaster, for callers that pace themselves without going through a full
// Clock (e.g. gatecommon's no-loader wall-clock fallbacks). Call once per
// goroutine, same rule as Subscribe.
func NewTickChan() <-chan struct{} {
	return globalTickBroadcaster().Subscribe()
}

// ApplySpeedNonBlocking is the delivery half of per-goroutine-clock.md
// "Delivery": every paced loop grows exactly this one poll, folded into its
// existing sleep/select point. speedCh is a buffered-1, latest-wins channel
// built once at load time (see loader.go / builders.go) and owned from then on
// by exactly the one goroutine that reads it — nothing else may read it.
// A pending value (if any) is drained and applied to clk's OWN
// copy via SetSpeed; an empty channel is a no-op; a nil channel (unwired
// goroutines, or test builds constructed with no loader) is always a no-op
// too, since a receive on a nil channel is never selected. This is
// non-blocking on purpose — a goroutine that is not yet awake must never be
// forced to wake early just to drain its inbox.
func ApplySpeedNonBlocking(clk Clock, speedCh <-chan float64) {
	select {
	case sp := <-speedCh:
		if rc, ok := clk.(*RealClock); ok {
			rc.SetSpeed(sp)
		}
	default:
	}
}

// SendSpeedNonBlocking is the send half of Delivery: it delivers speed to one
// clock-holder's buffered-1 channel WITHOUT blocking on a goroutine that may be
// asleep or never reads. If the buffer already holds a stale pending value
// (a rapid second change arrived before the holder woke to drain the first),
// that stale value is dropped and replaced — LATEST WINS is correct because
// speed is absolute state, not an event stream (per-goroutine-clock.md
// "Delivery"). ch must be a channel this call's caller alone sends on (the
// stdin-reader goroutine, which is the sole writer of every channel collected
// at load) — sending from two goroutines onto the same ch would race the
// drain-then-send pair below.
func SendSpeedNonBlocking(ch chan float64, speed float64) {
	select {
	case ch <- speed:
		return
	default:
	}
	// Buffer full: drain the stale value, then place the new one. Both steps are
	// non-blocking; if some other reader raced us and drained it first between
	// the two selects, the second send below still succeeds (buffer now empty).
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- speed:
	default:
	}
}

// SendLatestNonBlocking is the int64 twin of SendSpeedNonBlocking, used for
// delivering a "held" value from a node's main loop (the owner) to its own
// spawned DriveHeld goroutine(s) over a buffered-1, latest-wins channel — the
// same non-blocking drain-then-send shape, because held (like speed) is
// absolute state, not an event stream: a goroutine that wakes up cares only
// about the CURRENT held value, not every intermediate one it missed while
// asleep. ch must be a channel this call's caller alone sends on (the node's
// own main-loop goroutine, which owns the held value); sending from two
// goroutines onto the same ch would race the drain-then-send pair below. A
// node driving two outputs off one held value (e.g. Pulse's Out/OutFanout) must
// pass a DIFFERENT channel per DriveHeld goroutine and call this once per
// channel — passing the same channel to two DriveHeld goroutines would starve
// whichever one loses a given receive (exactly the speedCh rationale above).
func SendLatestNonBlocking(ch chan int64, v int64) {
	select {
	case ch <- v:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- v:
	default:
	}
}

// inertClock is GONE (per-goroutine-clock.md API demolition item 3). It existed only
// because an INJECTED clock could be ABSENT: an unwired In needed a non-nil thing to
// return from a port accessor, and the retired type-matched field injection meant a
// rename could silently inject nothing, leaving an unguarded clk.Tick() to panic with no
// recover over the node goroutine. A goroutine that constructs (or Copies) its own clock
// cannot have a nil one — every clock-holder now gets a real *RealClock, seeded from the
// loader's origin at construction, so there is no "absent" case left to paper over. Do
// not reintroduce a placeholder/inert Clock implementation; that is exactly the
// unrepresentable-nil trap this deletion removes.
