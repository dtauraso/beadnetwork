package Node1

// parallel.go — THE PARALLEL STATE MACHINE, and nothing else.
//
// A pair is parallel when its two tilts lie on the SAME LINE. Reversing one arrow does not change
// that — two arrows on one line are parallel either way round — so both are this machine's halt,
// and it reads them as one thing without being told to: `separation` folds the long way round
// into the short one, so a reversed partner lands exactly where an aligned one does. Swept over
// every opening of a 24-point lattice, the two look equally often — 264 each of 528 (math.html),
// which is worth knowing only because the arrows LOOK different in the editor. This file is the
// whole rule for
// that and knows no other resting state: nothing here names a zero or half-turn separation,
// nothing here asks what perpendicular is doing, and nothing here reads perpendicular.go. A
// zero separation is an ordinary angle to this machine — one it walks over on the way to its
// own halt.
//
// See perpendicular.go's header for why the two are separate files. This one is here to be
// LEFT ALONE while that one changes.
//
// WHAT ARRIVES is the partner's coplanar NORMAL, already a quarter turn off the partner's own
// tilt. So "the tilts are on the same line" reaches this machine as the arrival sitting a
// quarter turn off this node's own TOP — the partner's own quarter, read back, from whichever
// of the two directions the partner is pointing.

import "github.com/dtauraso/wirefold/nodes/Wiring"

type parallelMachine struct{}

// halted reports whether this arrival IS this machine's resting state.
func (parallelMachine) halted(from, arrival *tiltState) bool {
	return from.separation(arrival) == from.ring.quarterTurn
}

// step is the single move — next or prev, a link either way, so it cannot leave the ring —
// that leaves the node closer to this machine's halt.
func (m parallelMachine) step(from, arrival *tiltState) *tiltState {
	if m.miss(from.next, arrival) <= m.miss(from.prev, arrival) {
		return from.next
	}
	return from.prev
}

// miss is how far this arrival is from this machine's halt — zero when it is it.
func (parallelMachine) miss(from, arrival *tiltState) int32 {
	return abs32(from.separation(arrival) - from.ring.quarterTurn)
}

func (parallelMachine) String() string { return "parallel" }

func (parallelMachine) choice() Wiring.TiltMachine { return Wiring.TiltMachineParallel }
