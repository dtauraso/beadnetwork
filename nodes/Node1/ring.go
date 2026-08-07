package Node1

// ring.go — the θ lattice as a RING OF STATES rather than a number that gets arithmetic done
// to it, and one ring PER NODE rather than one for the process.
//
// A tilt can point in exactly as many directions as the lattice has points, so there are
// exactly that many states, and a tilt IS one of them. Turning one step is following a link:
//
//	n.Top = n.Top.next        // instead of idx = idx + 1, then wrap
//	n.Top = n.Top.prev        // instead of idx = idx - 1, then wrap
//
// WHAT THIS BUYS. Keeping an index in range means applying a wrap at every site that moves
// it — correct, but a convention rather than a structure: a site that forgot it produced an
// index a full turn from where it read, silently, and only a test walking the whole circle
// would notice. A state has no arithmetic to forget, and there is no state off the ring to
// point at, so that cannot be written down at all
// (memory/feedback_make_bug_class_unrepresentable.md).
//
// WHY ONE RING PER NODE. The point count is a scene setting a user can change, so the lattice
// is not fixed for the life of the process. A single package-level ring rebuilt on that
// change would be shared MUTABLE state read by every pair node's goroutine — the one thing
// this network does not do (CLAUDE.md's ownership rule; guard: check-no-network-locks). So
// each node builds its OWN ring, and a count change is that node building a new one on its
// own goroutine. A ring is immutable once built: nothing writes to one after newRing returns,
// so the states inside it need no protection and none is taken.
//
// Every relation the pair rule needs is a link, resolved once when the ring is built:
//
//	next / prev    one step either way
//	opposite       the bottom tilt, a half turn away
//	quarter        the coplanar normal, a quarter turn on
//
// so bottomTilt and coplanarNormal are field reads, not sums.

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// tiltState is one of its ring's directions. Values are the ring's own elements — a
// *tiltState is always one of them, never a fresh one, so pointer identity IS direction
// equality, WITHIN one ring. Two rings built for the same count hold equal but distinct
// states, which is why a node compares only against states from its own.
type tiltState struct {
	// idx is this state's own index, 0…points-1. READ ONLY, and it exists for the BOUNDARY
	// alone: the buffer column, position.json, and the vector-channel message all carry a
	// number. Nothing inside this kind computes with it except acuteWith's gap.
	idx int32

	// ring is the lattice this state belongs to — read for the two bounds acuteWith needs,
	// and what makes a state self-describing rather than needing its ring passed alongside.
	ring *ring

	next     *tiltState // one step on
	prev     *tiltState // one step back
	opposite *tiltState // a half turn away — this state's own bottom tilt
	quarter  *tiltState // a quarter turn on — this state's own coplanar normal
}

// ring is ONE NODE'S OWN lattice: its states, and the two counts every rule reads.
type ring struct {
	// points is how many directions this lattice has — the number the user sets, and the
	// number of states. quarterTurn and halfTurn are derived once here rather than at each
	// read: a quarter turn is points/4 and a half turn is points/2, both exact because
	// newRing refuses a count that is not a multiple of four.
	points      int32
	quarterTurn int32
	halfTurn    int32

	states []tiltState
}

// newRing builds a lattice of the given number of points, with every link resolved.
//
// THE COUNT MUST BE A POSITIVE MULTIPLE OF FOUR. A quarter turn has to be a whole number of
// states, because the coplanar normal is exactly one and the halt is exactly one: at a count
// of 25 there is no state a quarter turn from a given state, so "perpendicular" — the
// condition the exchange comes to rest on — names nothing. A count that cannot express the
// rule is not a lattice this kind can run on, so this refuses it rather than rounding to one
// that works and leaving the drawn result to explain itself.
func newRing(points int32) *ring {
	if points < 4 || points%4 != 0 {
		panic(fmt.Sprintf(
			"Node1: a lattice needs a positive multiple of four points — got %d; a quarter turn must be a whole number of states or the coplanar normal and the perpendicular halt name nothing",
			points))
	}
	r := &ring{
		points:      points,
		quarterTurn: points / 4,
		halfTurn:    points / 2,
		states:      make([]tiltState, points),
	}
	for i := range r.states {
		r.states[i].idx = int32(i)
		r.states[i].ring = r
	}
	n := int32(len(r.states))
	for i := int32(0); i < n; i++ {
		r.states[i].next = &r.states[(i+1)%n]
		r.states[i].prev = &r.states[(i-1+n)%n]
		r.states[i].opposite = &r.states[(i+r.halfTurn)%n]
		r.states[i].quarter = &r.states[(i+r.quarterTurn)%n]
	}
	return r
}

// at is the state with this index — how anything inside this kind names a direction it
// already knows this ring has. An index outside the ring is Go's own out-of-range panic,
// which is the correct outcome: there is no such direction.
func (r *ring) at(idx int32) *tiltState { return &r.states[idx] }

// arrivedState maps a direction ARRIVING ON THE VECTOR CHANNEL onto this ring's state.
//
// It does not reduce, because there is nothing legitimate to reduce. The sender is this same
// kind, sending one of its own states' idx (outgoingVector → quarter.idx), so an arrival is
// on the ring or something is wrong. Folding an out-of-range value would turn that into a
// direction a full turn from the one sent — plausible, drawable, and silent, which is the
// failure the ring exists to make impossible. So it panics instead, and names what was
// violated.
//
// A PARTNER AT A DIFFERENT POINT COUNT REACHES THIS, and should: the two ends of a pair
// index the same lattice, so a direction from a lattice of another size is not a direction
// this node can act on. Stopping is better than turning toward a number that means something
// else.
func (r *ring) arrivedState(idx int32) *tiltState {
	if idx < 0 || idx >= r.points {
		panic(fmt.Sprintf(
			"Node1: a direction arriving on the vector channel must already be an index on this node's own %d-point ring (0..%d) — got %d; the sender is this same kind sending one of its own states, so an index off the ring is a defect or a partner on a different lattice, not something to fold onto this one",
			r.points, r.points-1, idx))
	}
	return r.at(idx)
}

// seedState maps a PERSISTED index onto this ring by asking which state carries it — not by
// computing one. Every state was given its own index when the ring was built, so the ring is
// the authority on which indices exist, and a number either names one of them or names
// nothing.
//
// A number that names nothing loads at the origin, and reports that it did. The alternative
// is to fold it — 30 becoming 6 on a 24-point ring, a quarter turn from where the file said —
// which is a direction nobody wrote, arrived at by arithmetic, and indistinguishable once
// drawn from one somebody chose. Landing at the origin is wrong in an obvious way instead of
// a plausible one, and the caller says which number was refused.
//
// THIS IS ROUTINE, not legacy handling: a scene saved at one point count and opened at a
// smaller one names indices the new ring does not have, and so does a file written before the
// count was settable at all.
func (r *ring) seedState(idx int32) (s *tiltState, unknown bool) {
	for i := range r.states {
		if r.states[i].idx == idx {
			return &r.states[i], false
		}
	}
	return &r.states[0], true
}

// THE ACUTE TEST IS GONE. It asked whether the arrival lay within a quarter turn, and a cone
// says that without saying which SIDE, so it could not answer at exactly a quarter turn at all
// — it reported "not acute" there, which the rule read as "stand still", at precisely the
// separation a node holding perpendicular most needed to move off. Direction now comes from
// which single step leaves this node nearer its own halt (stepToward), which is answerable
// everywhere on the ring, including at both halts.
//
// separation is how far apart two states are on the ring, going the SHORT way round — never
// more than a half turn. acuteWith above answers a yes/no with the long-way case folded into
// its second comparison; the halt tests need the number itself, and need it to mean the same
// thing whichever side the target sits on, so the fold happens here instead.
func (s *tiltState) separation(target *tiltState) int32 {
	gap := s.idx - target.idx
	if gap < 0 {
		gap = -gap
	}
	if gap > s.ring.halfTurn {
		gap = s.ring.points - gap
	}
	return gap
}

// PERPENDICULAR AND PARALLEL ARE DIFFERENT STATES, AND EACH HAS ITS OWN HALT.
//
// What arrives is the partner's coplanar NORMAL, which already sits a quarter turn off the
// partner's own tilt. So the separation between this node's top and that arrival says what
// the two TILTS are doing, one quarter turn removed:
//
//	separation 0, or a half turn  ->  the tilts are a quarter turn apart  ->  PERPENDICULAR
//	separation a quarter turn     ->  the tilts are the same direction    ->  PARALLEL
//
// Both are places the pair can rest, and they are NOT the same place. The rule used to halt on
// "not acute", which is one condition covering both — so a pair disturbed out of perpendicular
// could walk into parallel and stop there, and the log read identically at both (every row
// `kind=none -> hold`). A node now names which of the two it is holding, because the direction
// it must turn when something disturbs it depends on which one it is trying to get back to.
type haltKind int8

const (
	haltNone haltKind = iota
	haltPerpendicular
	haltParallel
)

// String names the halt for the diagnostic row — the two have to be distinguishable there
// too, since a log that printed both as "halted" is what hid them being one state.
func (h haltKind) String() string {
	switch h {
	case haltPerpendicular:
		return "perpendicular"
	case haltParallel:
		return "parallel"
	}
	return "none"
}

// missBy is how far this arrival is from putting the node in the halt h — zero when it IS
// that halt. Perpendicular has two separations that are it, a zero and a half turn, so the
// nearer of the two is the one that counts; parallel has the one, a quarter turn.
//
// This is what lets a node STEP THROUGH the halt it is not holding. The walk back to one halt
// passes across the other — perpendicular sits at separation 0 and a half turn, parallel at a
// quarter turn, so getting from one to the other crosses the space between them, and the
// crossing point IS the other halt. A node that stopped at any halt was captured by whichever
// it touched first: the log showed a pair holding perpendicular walk correctly toward
// separation 0, land on 12 in passing, and take up parallel there. Measuring the miss against
// THIS NODE'S OWN halt makes the other one an ordinary angle it passes over.
// PARALLEL IS NOT CONSULTED UNLESS A NODE IS HOLDING IT. A node holding nothing yet closes on
// the arrival, which is the perpendicular measure — it needs no comparison against the other
// halt to do that, and a version that compared the two to pick a target made a node's distance
// to parallel part of answering a question about perpendicular.
func (s *tiltState) missBy(arrival *tiltState, h haltKind) int32 {
	sep := s.separation(arrival)
	if h == haltParallel {
		return abs32(sep - s.ring.quarterTurn)
	}
	toZero, toHalf := sep, abs32(sep-s.ring.halfTurn)
	if toHalf < toZero {
		return toHalf
	}
	return toZero
}

// stepToward is the one step — next or prev, a link either way — that leaves this node closer
// to its OWN halt against this arrival. It replaces picking a direction from the acute cones,
// which could not answer at all in one place it had to: a node holding perpendicular that has
// arrived at exactly a quarter turn is not acute on either side, so the cones said "stand
// still" at precisely the separation it most needed to move off.
func (s *tiltState) stepToward(arrival *tiltState, h haltKind) *tiltState {
	if s.next.missBy(arrival, h) <= s.prev.missBy(arrival, h) {
		return s.next
	}
	return s.prev
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// haltAgainst names the halt this arrival puts the node in, or haltNone when the arrival is
// somewhere between the two and the node has to turn.
func (s *tiltState) haltAgainst(arrival *tiltState) haltKind {
	switch s.separation(arrival) {
	case 0, s.ring.halfTurn:
		return haltPerpendicular
	case s.ring.quarterTurn:
		return haltParallel
	}
	return haltNone
}

// defaultRing is the lattice a node gets when nothing has said otherwise — the count this
// model has always run at, and what a bare test build in this package constructs against. It
// is built once and never written to, so it is immutable shared data rather than shared
// mutable state; a node given a different count builds its own ring instead of touching this.
var defaultRing = newRing(Wiring.FullTurnThetaIdx)
