// periodic_emit.go — Input's plain periodic-source emit path, entered from Update only
// when FeedbackIn is not wired. Grouped together because inputCadenceTicks is the one
// field read runPeriodicEmit's fire-cadence test calls every pass.

package input

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// runPeriodicEmit is Update's plain emit path (FeedbackIn not wired): Input is
// a periodic SOURCE. It pops the end and fans the value to every wired output
// (2 and 3), then sleeps ONE CADENCE — a sleep timer of (one human cycle) ×
// (the broadcast edge length) — before firing the next value. The bead is
// stepped one position per human cycle DURING that sleep, so it traverses the
// edge across the cadence; with equal-length output edges (assumed) both
// outputs stay in lockstep. With Repeat the buffer refills forever; without
// it, once the working buffer is drained it simply idles (no fire) but keeps
// cycling. Layout/drag handling is NOT here — the node's dedicated always-on
// layout goroutine owns it (split-layout-bead-goroutines.md), independent of
// this pausable bead loop. clk is the same copy Update took once at startup;
// no second derivation.
func (n *Node) runPeriodicEmit(ctx context.Context, working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) {
	emitted := 0
	// Fire cadence is measured in CLOCK TICKS, exactly like a gate's window/dwell
	// (gatecommon/gate.go: fire when now()-dwellStart >= fireDwellTicks). Tick()
	// freezes on Halt, so the cadence — and therefore emission — freezes on pause
	// just like every other node kind. The multiplication factor is the only
	// Input-specific part: the cadence is one tick per unit of the broadcast edge
	// length, recomputed each pass so a drag re-paces it.
	lastFireTick := clk.Tick() - int64(inputCadenceTicks(n)) // fire on the first pass
	for {
		if ctx.Err() != nil {
			return
		}
		now := clk.Tick()
		if (n.Repeat || emitted < len(init)) && now-lastFireTick >= int64(inputCadenceTicks(n)) {
			if n.Fire != nil {
				n.Fire()
			}
			v := popEnd(working, backup, init)
			emitBeads() // array changed (pop, maybe refill) → restream interior
			if !n.broadcastPlace(v, now) {
				return
			}
			lastFireTick = now
			emitted++
		}

		clock.ApplySpeedNonBlocking(clk, n.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

// inputCadenceTicks is Input's fire cadence in clock ticks: the CROSSING TIME of
// the OutCadence edge, Steps * DwellTicksPerBead (= ticksToCross,
// docs/bead-model/bead-lattice.md "Timing"), so exactly one bead crosses the edge per
// cadence — no overlap. Measured in ticks, so it freezes on pause with Tick().
// Recomputed live so a drag that changes the edge's step count re-paces emission. The
// arithmetic itself is cadenceTicks (emit_helpers.go); this wrapper's only job is the
// one field read (n.OutCadence.Geom().Steps) that arithmetic can't do itself.
func inputCadenceTicks(n *Node) int64 {
	return cadenceTicks(n.OutCadence.Geom().Steps)
}
