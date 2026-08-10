// arrival.go — ONE JOB: the arrival math. When does a bead of a given step count
// finish crossing (ticksToCross), when does THIS wire's earliest bead land
// (NextArrivalTick), and when does the earliest bead across a set of wires land
// (EarliestArrival). Pure reads over already-placed beads: nothing here places,
// advances or delivers.

package wire

import "math"

// ticksToCross returns the tick count for a bead of the given STEP count to
// cross, at this wire's own dwell-per-bead: steps * dwell (docs/bead-model/bead-lattice.md
// "Timing" — a longer edge is simply more beads, dwell is a constant, so there
// is no per-edge division left the way arcLength/pulseSpeed used to require).
// Fractional; the driver delivers on the first integer tick at or past
// placementTick + this.
func (pw *PacedWire) ticksToCross(steps int) float64 {
	if pw.dwell <= 0 {
		return 0
	}
	return float64(steps) * pw.dwell
}

// NextArrivalTick is the tick the EARLIEST in-flight bead on this wire will finish
// crossing on — known at PLACEMENT, not discovered by watching: a bead placed at
// placementTick with a step count of steps lands at placementTick + steps*dwell, and
// nothing about that changes while it flies. ok is false when nothing is in flight.
//
// This is what lets the source node SLEEP to the moment a traversal completes instead of
// waking every cycle to ask whether it has. The animation is unaffected either way —
// LiveBeadFractions is a pure function of (current tick, placementTick), so the pulse's
// drawn position never depended on this goroutine being awake.
//
// Called on this wire's own goroutine, like every other reader of inflight.
func (pw *PacedWire) NextArrivalTick() (int64, bool) {
	best := int64(0)
	found := false
	for _, b := range pw.inflight {
		// Ceil: delivery happens on the first INTEGER tick at or past the deadline
		// (ticksToCross' own doc comment), so the wake must not land a fraction early.
		at := int64(math.Ceil(b.placementTick + pw.ticksToCross(b.steps)))
		if !found || at < best {
			best, found = at, true
		}
	}
	return best, found
}

// EarliestArrival is the SOONEST arrival across n wires — what a node with several
// outgoing edges sleeps to. The shorter edge wins because it is the first thing that needs
// this goroutine awake; a longer one simply has not arrived yet, and waking for it early is
// the polling this replaces. ok is false when nothing is in flight on any of them.
//
// A free function over a slice rather than a method on the node: the node owns which wires
// it drives, and this owns none of them — it only reads each one's own answer.
func EarliestArrival(wires []*PacedWire) (int64, bool) {
	best := int64(0)
	found := false
	for _, w := range wires {
		if w == nil {
			continue
		}
		if at, ok := w.NextArrivalTick(); ok && (!found || at < best) {
			best, found = at, true
		}
	}
	return best, found
}
