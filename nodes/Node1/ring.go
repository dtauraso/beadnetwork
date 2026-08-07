package Node1

// ring.go — the θ lattice as a RING OF STATES rather than a number that gets arithmetic done
// to it.
//
// There are exactly Wiring.FullTurnThetaIdx directions a tilt can point, so there are exactly
// that many states, and a tilt IS one of them. Turning one step is following a link:
//
//	n.top = n.top.next        // instead of idx = idx + 1, then wrap
//	n.top = n.top.prev        // instead of idx = idx - 1, then wrap
//
// WHAT THIS BUYS. The previous form kept the index on the circle by applying a wrap at every
// site that moved it — correct, but a convention rather than a structure: a site that forgot
// it produced an index 24 steps from where it read, silently, and only a test walking the
// whole circle would notice. A state has no arithmetic to forget. There is no twenty-fifth
// state to point at, so "off the circle" cannot be written down at all
// (memory/feedback_make_bug_class_unrepresentable.md).
//
// Every relation the pair rule needs is a link, resolved once when the ring is built:
//
//	next / prev    one step either way
//	opposite       the bottom tilt, a half turn away
//	quarter        the coplanar normal, a quarter turn on
//
// so bottomTilt and coplanarNormal are field reads, not sums.
//
// WHAT ARITHMETIC IS LEFT, and where:
//
//   - Two boundary functions, where a NUMBER has to become a state, because the two callers
//     want different things from an out-of-range value: arrivedState, for a direction ARRIVING
//     ON THE VECTOR CHANNEL — the sender is this same kind, so an out-of-range index is a
//     defect, and it panics rather than fold one. seedState, for the PERSISTED SEED — an older
//     build's file can legitimately hold one, so it folds onto the ring (the only `%` in this
//     kind) and reports whether it had to.
//   - acuteWith, which subtracts two states' own indices to get the gap between them. It
//     needs no reduction of any kind, because both are on the ring: larger minus smaller is
//     already the gap, and the two bounds it is tested against cover both ways round.
//
// Neither can put a tilt somewhere it cannot be — a tilt only ever moves by following a link.

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// at is the ring member with this index — the primitive both boundary functions below are
// built from, and how anything inside this kind names a direction it already knows is on the
// ring. An index outside the ring is Go's own out-of-range panic, which is the correct
// outcome: there is no such direction.
func at(idx int32) *tiltState { return &ring[idx] }

// tiltState is one of the FullTurnThetaIdx directions. Values are the ring elements
// themselves — a *tiltState is always one of them, never a fresh one, so pointer identity IS
// direction equality.
type tiltState struct {
	// idx is this state's own index, 0…FullTurnThetaIdx-1. READ ONLY, and it exists for the
	// BOUNDARY alone: the buffer column, position.json, and the vector-channel message all
	// carry a number. Nothing inside this kind computes with it.
	idx int32

	next     *tiltState // one step on
	prev     *tiltState // one step back
	opposite *tiltState // a half turn away — this state's own bottom tilt
	quarter  *tiltState // a quarter turn on — this state's own coplanar normal
}

// ring is the whole state space, indexed by direction. Built once at package init and never
// written again, so every goroutine reads it freely — it is immutable shared data, not shared
// mutable state (CLAUDE.md's ownership rule is about the latter).
var ring [Wiring.FullTurnThetaIdx]tiltState

func init() {
	for i := range ring {
		ring[i].idx = int32(i)
	}
	for i := range ring {
		ring[i].next = &ring[(i+1)%len(ring)]
		ring[i].prev = &ring[(i-1+len(ring))%len(ring)]
		ring[i].opposite = &ring[(i+int(Wiring.HalfTurnThetaIdx))%len(ring)]
		ring[i].quarter = &ring[(i+int(Wiring.PerpendicularThetaIdx))%len(ring)]
	}
}

// arrivedState maps a direction ARRIVING ON THE VECTOR CHANNEL onto its state.
//
// It does not reduce, because there is nothing legitimate to reduce. The sender is this same
// kind, sending one of its own ring members' idx (outgoingVector → quarter.idx), so an
// arrival is on the ring or the program is wrong. Folding an out-of-range value would turn
// that bug into a direction 24 steps from the one that was sent — plausible, drawable, and
// silent, which is the failure the ring exists to make impossible. So it panics instead, and
// names what was violated.
//
// The panic is reachable only from a defect in this package or a foreign writer on a channel
// only this kind holds; a partner at a different lattice size would reach it too, and that is
// worth stopping on rather than rendering.
func arrivedState(idx int32) *tiltState {
	if idx < 0 || idx >= Wiring.FullTurnThetaIdx {
		panic(fmt.Sprintf(
			"Node1: a direction arriving on the vector channel must already be a ring index in 0..%d — got %d; the sender is this same kind sending one of its own states, so an index off the ring is a defect, not something to fold onto the ring",
			Wiring.FullTurnThetaIdx-1, idx))
	}
	return at(idx)
}

// seedState maps the PERSISTED SEED onto its state by ASKING THE RING which state carries
// that index — not by computing one. Every state was given its own index when the ring was
// built, so the ring is the authority on which indices exist, and a number either names one
// of them or names nothing.
//
// A number that names nothing loads as the ring's origin, and reports that it did. The
// alternative is to fold it — 30 becoming 6, a quarter turn from where the file said — which
// is a direction nobody wrote, arrived at by arithmetic, and indistinguishable once drawn
// from one somebody chose. Landing at the origin is wrong in an obvious way instead of a
// plausible one, and the breadcrumb says which number was refused.
//
// This is the case a persisted file legitimately reaches: position.json is written by an
// older build as readily as this one, and once the lattice size is editable, a file saved at
// one size and opened at a smaller one names indices this ring does not have.
//
// There is no arithmetic here at all — no modulo, no sign to handle. It is 24 comparisons
// once per node at load.
func seedState(idx int32) (s *tiltState, folded bool) {
	for i := range ring {
		if ring[i].idx == idx {
			return &ring[i], false
		}
	}
	return &ring[0], true
}

// acuteWith reports whether target lies within a quarter turn of s — the whole of what the
// straightening rule asks about two directions.
//
// THE GAP IS TAKEN LARGER MINUS SMALLER, so it is never negative and there is no sign
// convention anywhere: both states are on the ring, so the gap lands in [0, FullTurnThetaIdx)
// with no reduction of any kind — no modulo, no conditional add, no floor.
//
// A gap and its complement describe the same pair of directions, one going each way round the
// ring, and the shorter of the two is the angle between them. So rather than pick the shorter
// — a min, and another comparison — both are tested at once: the pair is within a quarter
// turn when the gap is under PerpendicularThetaIdx going one way, or over
// FullTurnThetaIdx-PerpendicularThetaIdx, which is the same thing going the other. Exactly a
// quarter turn (the gap at either bound) is NOT acute, and that is the perpendicular case the
// exchange halts on.
func (s *tiltState) acuteWith(target *tiltState) bool {
	gap := s.idx - target.idx
	if target.idx > s.idx {
		gap = target.idx - s.idx
	}
	return gap < Wiring.PerpendicularThetaIdx || gap > Wiring.FullTurnThetaIdx-Wiring.PerpendicularThetaIdx
}
