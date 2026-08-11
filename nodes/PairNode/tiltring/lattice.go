// Package tiltring is the θ lattice PairNode's tilts live on, and the one state machine that
// decides where a tilt returns to — pure math, no node, no goroutine, no channel. It was
// ring.go/machine.go inside package PairNode; the pure lattice/machine arithmetic and the tests
// that exercise it (arith_*_test.go, ring_test.go, machine_halt_test.go) moved here together, so
// the math sits with its own proofs. Everything that reads or writes a *PairNode.Node — which end
// an update names, which mode a node adopts, the drain/adopt loop — stayed in package PairNode,
// since that is state a node owns, not a fact about the ring.
//
// A tilt can point in exactly as many directions as the lattice has points, so there are exactly
// that many states, and a tilt IS one of them. Turning one step is following a link:
//
//	top = top.Next        // instead of idx = idx + 1, then wrap
//	top = top.Prev        // instead of idx = idx - 1, then wrap
//
// WHAT THIS BUYS. Keeping an index in range means applying a wrap at every site that moves it —
// correct, but a convention rather than a structure: a site that forgot it produced an index a
// full turn from where it read, silently, and only a test walking the whole circle would notice.
// A state has no arithmetic to forget, and there is no state off the ring to point at, so that
// cannot be written down at all (memory/feedback_make_bug_class_unrepresentable.md).
//
// WHY ONE RING PER NODE. The point count is a scene setting a user can change, so the lattice is
// not fixed for the life of the process. A single package-level ring rebuilt on that change would
// be shared MUTABLE state read by every pair node's goroutine — the one thing this network does
// not do (CLAUDE.md's ownership rule; guard: check-no-network-locks). So each node builds its OWN
// ring (PairNode.Node.adoptLattice), and a count change is that node building a new one on its own
// goroutine. A Ring is immutable once built: nothing writes to one after NewRing returns, so the
// states inside it need no protection and none is taken.
//
// Every relation the pair rule needs is a link, resolved once when the ring is built:
//
//	Next / Prev    one step either way
//	Opposite       the bottom tilt, a half turn away
//	Quarter        the coplanar normal, a quarter turn on
//
// so a node's bottom tilt and coplanar normal are field reads, not sums.
package tiltring

import "fmt"

// State is one of its ring's directions. Values are the ring's own elements — a *State is always
// one of them, never a fresh one, so pointer identity IS direction equality, WITHIN one ring. Two
// rings built for the same count hold equal but distinct states, which is why a caller compares
// only against states from its own.
type State struct {
	// Idx is this state's own index, 0…points-1. It exists for the BOUNDARY: the buffer column,
	// position.json, and the vector-channel message all carry a number.
	Idx int32

	// R is the lattice this state belongs to — read for the two bounds AngleLength needs, and
	// what makes a state self-describing rather than needing its ring passed alongside.
	R *Ring

	Next     *State // one step on
	Prev     *State // one step back
	Opposite *State // a half turn away — this state's own bottom tilt
	Quarter  *State // a quarter turn on — this state's own coplanar normal
}

// Ring is one lattice: its states, and the two counts every rule reads.
type Ring struct {
	// Points is how many directions this lattice has — the number the user sets, and the number
	// of states. QuarterTurn and HalfTurn are derived once here rather than at each read: a
	// quarter turn is points/4 and a half turn is points/2, both exact because NewRing refuses a
	// count that is not a multiple of four.
	Points      int32
	QuarterTurn int32
	HalfTurn    int32

	States []State
}

// NewRing builds a lattice of the given number of points, with every link resolved.
//
// THE COUNT MUST BE A POSITIVE MULTIPLE OF FOUR. A quarter turn has to be a whole number of
// states, because the coplanar normal is exactly one and the halt is exactly one: at a count of
// 25 there is no state a quarter turn from a given state, so "perpendicular" — the condition the
// exchange comes to rest on — names nothing. A count that cannot express the rule is not a
// lattice this kind can run on, so this refuses it rather than rounding to one that works and
// leaving the drawn result to explain itself.
func NewRing(points int32) *Ring {
	if points < 4 || points%4 != 0 {
		panic(fmt.Sprintf(
			"tiltring: a lattice needs a positive multiple of four points — got %d; a quarter turn must be a whole number of states or the coplanar normal and the perpendicular halt name nothing",
			points))
	}
	r := &Ring{
		Points:      points,
		QuarterTurn: points / 4,
		HalfTurn:    points / 2,
		States:      make([]State, points),
	}
	for i := range r.States {
		r.States[i].Idx = int32(i)
		r.States[i].R = r
	}
	n := int32(len(r.States))
	for i := int32(0); i < n; i++ {
		r.States[i].Next = &r.States[(i+1)%n]
		r.States[i].Prev = &r.States[(i-1+n)%n]
		r.States[i].Opposite = &r.States[(i+r.HalfTurn)%n]
		r.States[i].Quarter = &r.States[(i+r.QuarterTurn)%n]
	}
	return r
}

// At is the state with this index — how anything holding a ring names a direction it already
// knows the ring has. An index outside the ring is Go's own out-of-range panic, which is the
// correct outcome: there is no such direction.
func (r *Ring) At(idx int32) *State { return &r.States[idx] }

// ArrivedState maps a direction ARRIVING ON THE VECTOR CHANNEL onto this ring's state.
//
// It does not reduce, because there is nothing legitimate to reduce. The sender is this same
// kind, sending one of its own states' Idx (PairNode's outgoingVector → Quarter.Idx), so an
// arrival is on the ring or something is wrong. Folding an out-of-range value would turn that
// into a direction a full turn from the one sent — plausible, drawable, and silent, which is the
// failure the ring exists to make impossible. So it panics instead, and names what was violated.
//
// A PARTNER AT A DIFFERENT POINT COUNT REACHES THIS, and should: the two ends of a pair index the
// same lattice, so a direction from a lattice of another size is not a direction this node can
// act on. Stopping is better than turning toward a number that means something else.
func (r *Ring) ArrivedState(idx int32) *State {
	if idx < 0 || idx >= r.Points {
		panic(fmt.Sprintf(
			"tiltring: a direction arriving on the vector channel must already be an index on this node's own %d-point ring (0..%d) — got %d; the sender is this same kind sending one of its own states, so an index off the ring is a defect or a partner on a different lattice, not something to fold onto this one",
			r.Points, r.Points-1, idx))
	}
	return r.At(idx)
}

// SeedState maps a PERSISTED index onto this ring by asking which state carries it — not by
// computing one. Every state was given its own index when the ring was built, so the ring is the
// authority on which indices exist, and a number either names one of them or names nothing.
//
// A number that names nothing loads at the origin, and reports that it did. The alternative is to
// fold it — 30 becoming 6 on a 24-point ring, a quarter turn from where the file said — which is
// a direction nobody wrote, arrived at by arithmetic, and indistinguishable once drawn from one
// somebody chose. Landing at the origin is wrong in an obvious way instead of a plausible one, and
// the caller says which number was refused.
//
// THIS IS ROUTINE, not legacy handling: a scene saved at one point count and opened at a smaller
// one names indices the new ring does not have, and so does a file written before the count was
// settable at all.
func (r *Ring) SeedState(idx int32) (s *State, unknown bool) {
	for i := range r.States {
		if r.States[i].Idx == idx {
			return &r.States[i], false
		}
	}
	return &r.States[0], true
}

// AngleLength is how far apart two states are on the ring, going the SHORT way round — never
// more than a half turn.
func (s *State) AngleLength(target *State) int32 {
	// TWO POINTS ON A CIRCLE HAVE TWO ARCS BETWEEN THEM, one each way round, and this is the
	// shorter. That choice is not wrap handling and is not something the ring's states already
	// answer: a STATE cannot leave the ring, which is why turning never does arithmetic, but a
	// MEASUREMENT between two states still has to say which of the two arcs it means.
	//
	// One subtraction and a sign test. NOT max-min: a comparison is itself a subtraction, so
	// ordering the operands first would subtract twice to avoid a sign check that the sign bit
	// already answers. Both indices are in [0, points), so no modulus is needed or taken — d is
	// simply one of the two arcs, and points-d is the other.
	d := s.Idx - target.Idx
	if d < 0 {
		d = -d
	}
	if d > s.R.HalfTurn {
		d = s.R.Points - d
	}
	return d
}
