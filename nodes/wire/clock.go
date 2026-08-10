// clock.go — the single human-speed clock the network reads.
//
// MODEL.md pins this: there is exactly one clock — the system monotonic clock
// read through a fixed SCALE so it advances in integer TICKS at human-watchable
// speed (`tick = ⌊(systemNow − start) × scale⌋`). All timing in the network is
// tick counts, never wall-clock durations: goroutines pace themselves with
// SleepCycle, which blocks for exactly one clock cycle. A bead crossing an edge
// takes `ticksToCross = steps * DwellTicksPerBead` ticks (steps: the edge's own
// bead-step count, docs/bead-model/bead-lattice.md "The count"; DwellTicksPerBead: the one
// uniform per-step dwell, bead_lattice.go); node
// processing windows are tick counts. There is no separate render cadence — the
// tick IS the animation clock.
//
// SleepCycle no longer blocks on a per-caller time.After: it receives from a dedicated
// channel subscribed against the process's ONE TickBroadcaster goroutine (tick_broadcaster.go),
// which is the only thing that ever waits on wall time. Every pacing loop still calls
// SleepCycle exactly the same way; only what it blocks on changed.
//
// SCALE arithmetic (behavior-preserving vs. the retired wall-clock model, and
// vs. the later retired arc-length model — docs/bead-model/bead-lattice.md superseded both):
// the original model sampled bead positions every 16 ms. We pick one tick ≈ one
// old 16 ms sample:
//
//	MsPerTick = 16   ⇒   scale = 1 tick / 16 ms = 62.5 ticks/sec.
//
// So a bead visits ~the same number of positions in ~the same wall time, and
// pause/resume look identical.
//
// The model is sleep-only: pacing loops call SleepCycle to wait exactly one
// clock cycle rather than blocking on a target tick. RealClock (real_clock.go) is
// the single production Clock implementation; its pacing methods live in
// sleep_cycle.go, the process-wide tick pulse in tick_broadcaster.go, and the
// speed-delivery helpers in speed_delivery.go.
//
// The clock is free-running: there is no play/pause gate (that feature was removed
// end-to-end), so the tick advances monotonically with wall time for the life of
// the process.

package wire

import (
	"context"
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
	// SleepUntilTick blocks until Tick() reaches target (or ctx is done), by sleeping
	// whole cycles — it never waits on wall time directly, so a speed change mid-wait is
	// picked up on the next cycle exactly as a one-cycle loop would (and a clock held at
	// speed 0 waits forever, correctly: no ticks are accruing, so nothing arrives).
	//
	// This is what a node with beads in flight sleeps on: the earliest arrival tick is
	// known at placement (PacedWire.NextArrivalTick / EarliestArrival), so the goroutine
	// can wait for that moment instead of waking every cycle to ask whether it has come.
	// Returns immediately when target is already past.
	SleepUntilTick(ctx context.Context, target int64) error
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

// inertClock is GONE (per-goroutine-clock.md API demolition item 3). It existed only
// because an INJECTED clock could be ABSENT: an unwired In needed a non-nil thing to
// return from a port accessor, and the retired type-matched field injection meant a
// rename could silently inject nothing, leaving an unguarded clk.Tick() to panic with no
// recover over the node goroutine. A goroutine that constructs (or Copies) its own clock
// cannot have a nil one — every clock-holder now gets a real *RealClock, seeded from the
// loader's origin at construction, so there is no "absent" case left to paper over. Do
// not reintroduce a placeholder/inert Clock implementation; that is exactly the
// unrepresentable-nil trap this deletion removes.
