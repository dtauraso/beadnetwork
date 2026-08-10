package PairNode

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

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
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
			"PairNode: a lattice needs a positive multiple of four points — got %d; a quarter turn must be a whole number of states or the coplanar normal and the perpendicular halt name nothing",
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
			"PairNode: a direction arriving on the vector channel must already be an index on this node's own %d-point ring (0..%d) — got %d; the sender is this same kind sending one of its own states, so an index off the ring is a defect or a partner on a different lattice, not something to fold onto this one",
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
// angle length a node holding perpendicular most needed to move off. Direction now comes from
// which single step leaves this node nearer its own halt (stepToward), which is answerable
// everywhere on the ring, including at both halts.
//
// angle length is how far apart two states are on the ring, going the SHORT way round — never
// more than a half turn. acuteWith above answers a yes/no with the long-way case folded into
// its second comparison; the halt tests need the number itself, and need it to mean the same
// thing whichever side the target sits on, so the fold happens here instead.
func (s *tiltState) angleLength(target *tiltState) int32 {
	// TWO POINTS ON A CIRCLE HAVE TWO ARCS BETWEEN THEM, one each way round, and this is
	// the shorter. That choice is not wrap handling and is not something the ring's 24
	// states already answer: a STATE cannot leave the ring, which is why turning never
	// does arithmetic, but a MEASUREMENT between two states still has to say which of
	// the two arcs it means.
	//
	// One subtraction and a sign test. NOT max-min: a comparison is itself a subtraction,
	// so ordering the operands first would subtract twice to avoid a sign check that the
	// sign bit already answers. Both indices are in [0, points), so no modulus is needed
	// or taken — d is simply one of the two arcs, and points-d is the other.
	d := s.idx - target.idx
	if d < 0 {
		d = -d
	}
	if d > s.ring.halfTurn {
		d = s.ring.points - d
	}
	return d
}

// PERPENDICULAR AND PARALLEL ARE DIFFERENT STATES, AND EACH HAS ITS OWN HALT.
//
// What arrives is the partner's coplanar NORMAL, which already sits a quarter turn off the
// partner's own tilt. So the angle length between this node's top and that arrival says what
// the two TILTS are doing, one quarter turn removed:
//
//	angle length 0, or a half turn  ->  the tilts are a quarter turn apart  ->  PERPENDICULAR
//	angle length a quarter turn     ->  the tilts lie on one line          ->  PARALLEL
//
// Both are places the pair can rest, and they are NOT the same place. The rule used to halt on
// "not acute", which is one condition covering both — so a pair disturbed out of perpendicular
// could walk into parallel and stop there, and the log read identically at both (every row
// `kind=none -> hold`). A node now RUNS ONE OF TWO STATE MACHINES, and which one it is running
// is what says where it is returning to when something disturbs it.
//
// A node runs ONE MODE of the one machine (machine.go), or none yet. The modes differ only in
// which angle lengths they call home, and that difference is written as data — see machine.go's
// header for the audit that established it and for why the rule is now written once.
//
// THE RESTING-STATE RULES ARE NOT IN THIS FILE. They are the stopping counts in machine.go. What this
// file provides them is `angle length`: a measurement of where two directions sit relative to each
// other, which is not a rule and names no resting state.

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// defaultRing is the lattice a node gets when nothing has said otherwise — the count this
// model has always run at, and what a bare test build in this package constructs against. It
// is built once and never written to, so it is immutable shared data rather than shared
// mutable state; a node given a different count builds its own ring instead of touching this.
var defaultRing = newRing(tiltvector.FullTurnThetaIdx)

// A NODE'S OWN PLACE ON THE RING — the accessors every other file reads its lattice and its
// two ends through, and the one function that gives it a different lattice. They live here
// because each is about the ring rather than about the exchange: what a nil field falls back
// to, which end an update names, and what survives a change of point count.

// ringOf is this node's own lattice, with the default standing in for a Ring that was never
// set — a bare test build. Every read of the lattice goes through here.
func (n *Node) ringOf() *ring {
	if n.lattice.Ring == nil {
		return defaultRing
	}
	return n.lattice.Ring
}

// topState is this node's own tilt direction, with its ring's origin standing in for a Top
// that was never set — see the field's own doc comment. Every read of the tilt goes through
// here, so nothing else in this file has to care about that case.
func (n *Node) topState() *tiltState {
	if n.tilt.Top == nil {
		return n.ringOf().at(0)
	}
	return n.tilt.Top
}

// bottomState is the other end of the same line, read the same way topState reads the first.
func (n *Node) bottomState() *tiltState {
	if n.tilt.Bottom == nil {
		return n.topState().opposite
	}
	return n.tilt.Bottom
}

// setTop and setBottom are THE ONLY WAYS EITHER END IS WRITTEN, and each writes BOTH: the end
// named, and the other read straight off its opposite link in the same statement. That is what
// makes the two unable to disagree — not a rule to follow but the only spelling available, so a
// future update that drives one end cannot leave the other where it was.
//
// Which one a caller reaches for says which end its measurement was taken at, which is the whole
// reason both are stored (see the Bottom field).
func (n *Node) setTop(top *tiltState) { n.tilt.Top, n.tilt.Bottom = top, top.opposite }

func (n *Node) setBottom(bottom *tiltState) { n.tilt.Bottom, n.tilt.Top = bottom, bottom.opposite }

// fromAnotherLattice is the drop test for an arrival, and the reason it is a test rather than
// a fold is the whole of this file's argument about indices.
//
// A DIRECTION FROM ANOTHER LATTICE IS NOT A DIRECTION HERE. The two ends of a pair adopt
// a new point count at their own moments, each on its own goroutine, so between those
// moments an index picked on the old lattice can land here — where it names a different
// angle, or no state at all. Dropping it is the definite answer: the partner adopts the
// same count within its own next cycle and the exchange resumes from directions both
// ends can read. Zero is a bare test build that stated nothing, and is taken as this
// node's own lattice.
func (n *Node) fromAnotherLattice(received tiltvector.TiltVectorMsg) bool {
	return received.Points != 0 && received.Points != n.ringOf().points
}

// drainLattice drains LatticeIn non-blocking: a new point count for this node's own ring.
// Drained BEFORE the vector cycle (Update) so that anything already queued on VectorIn is
// discarded by the adopt rather than read one last time against the lattice it was not
// picked on.
func (n *Node) drainLattice() {
	if n.lattice.LatticeIn == nil {
		return
	}
	select {
	case points := <-n.lattice.LatticeIn:
		n.adoptLattice(points)
	default:
	}
}

// adoptLattice rebuilds THIS node's own ring at a new point count, on THIS node's own
// goroutine. Nothing else touches the ring, so there is nothing to coordinate: the old one is
// simply dropped and a new one takes its place.
//
// WHAT SURVIVES THE CHANGE IS THE INDEX, not the angle. A tilt at 6 stays at 6 — which is a
// quarter turn on a 24-point lattice and a half turn on a 12-point one, so the drawn arrow
// moves. That is the honest reading of "the lattice changed underneath a direction": the
// number a user set is kept, and what it means follows the new lattice. An index the new ring
// does not have names nothing there, so that node opens at the origin and says so
// (ring.seedState).
//
// TWO THINGS ARE DISCARDED, both because they are indices on the lattice being left:
//
//   - the received direction, the third drawn arrow. It was picked on the old lattice, so
//     redrawing it at the same index would point it somewhere the partner never sent.
//   - whatever is queued on VectorIn. Same reason, and it would otherwise be read as a
//     direction on the new ring the moment the next cycle polls.
//
// The beads in flight are untouched: a bead carries no direction, only pacing.
func (n *Node) adoptLattice(points int32) {
	if points == n.ringOf().points {
		return
	}
	keptIdx := n.topState().idx
	n.lattice.Ring = newRing(points)
	top, unknown := n.lattice.Ring.seedState(keptIdx)
	n.setTop(top)
	if unknown && n.plumb.Self != nil {
		n.plumb.Self.Breadcrumb("pair-lattice-adopt", fmt.Sprintf(
			"points=%d keptIdx=%d unknown=true loaded=%d", points, keptIdx, top.idx))
	}
	n.vec.ReceivedThetaIdx = 0
	n.vec.ReceivedSet = false
	n.syncReceivedVector()
	tiltvector.PollRecvVector(n.vec.VectorIn)
	if n.lattice.SyncLatticePoints != nil {
		n.lattice.SyncLatticePoints(points)
	}
	n.syncTiltIndex()
}
