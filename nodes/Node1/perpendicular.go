package Node1

// perpendicular.go — THE PERPENDICULAR STATE MACHINE, and nothing else.
//
// A pair is perpendicular when its two tilts are a quarter turn apart. This file is the whole
// rule for that and knows no other resting state: no test here names a quarter-turn separation,
// nothing here asks what parallel is doing, and nothing here reads parallel.go. A quarter turn
// is an ordinary angle to this machine — one it walks over on the way to its own halt.
//
// It is a separate file from parallel.go on purpose. Three attempts at this rule changed the
// parallel one as a side effect, every time through something that served both: a miss function
// parameterized by which state you were in, a step that took the state as an argument, and a
// version where a node holding neither compared its distance to parallel against its distance
// to perpendicular to choose a target. Two machines in two files have nowhere to put that
// coupling (memory/feedback_code_self_defends.md). They will look alike; that is the cost, and
// folding them back together is the thing this arrangement exists to prevent.
//
// WHAT ARRIVES is the partner's coplanar NORMAL, already a quarter turn off the partner's own
// tilt. So "the tilts are a quarter turn apart" reaches this machine as the arrival sitting
// exactly on this node's own TOP, or exactly on its BOTTOM — separation 0, or a half turn.

type perpendicularMachine struct{}

// halted reports whether this arrival IS this machine's resting state.
func (perpendicularMachine) halted(from, arrival *tiltState) bool {
	sep := from.separation(arrival)
	return sep == 0 || sep == from.ring.halfTurn
}

// step is the single move — next or prev, a link either way, so it cannot leave the ring —
// that leaves the node closer to this machine's halt. The halt has two separations that are
// it, and this closes on whichever is nearer.
func (m perpendicularMachine) step(from, arrival *tiltState) *tiltState {
	if m.miss(from.next, arrival) <= m.miss(from.prev, arrival) {
		return from.next
	}
	return from.prev
}

// miss is how far this arrival is from this machine's halt — zero when it is it.
func (perpendicularMachine) miss(from, arrival *tiltState) int32 {
	sep := from.separation(arrival)
	if toHalf := abs32(sep - from.ring.halfTurn); toHalf < sep {
		return toHalf
	}
	return sep
}

func (perpendicularMachine) String() string { return "perpendicular" }
