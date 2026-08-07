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
// THE ONE ARITHMETIC LEFT is at the edges, where a number has to become a state: the
// persisted seed and a direction arriving from the partner. Both go through stateFor, which
// reduces once, in one place, and is the only `%` in this kind.

import "github.com/dtauraso/wirefold/nodes/Wiring"

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

// stateFor maps a NUMBER from outside this kind onto its state — the persisted seed, and a
// direction arriving on the vector channel. This is the boundary, and the only place a
// modulo appears: an arriving partner's index is on the ring by construction, and a
// position.json written by an older build may hold anything, and neither is worth trusting
// over one reduction.
func stateFor(idx int32) *tiltState {
	i := ((idx % Wiring.FullTurnThetaIdx) + Wiring.FullTurnThetaIdx) % Wiring.FullTurnThetaIdx
	return &ring[i]
}

// acuteWith reports whether target lies within a quarter turn of s — the whole of what the
// straightening rule asks about two directions.
//
// STATED AS REACHABILITY, not as a difference: acute means "at most PerpendicularThetaIdx-1
// moves away, either way round", so this walks out from s in both directions and looks for
// target. Exactly a quarter turn away is NOT acute (the walk stops one short of it), which is
// the perpendicular case the exchange halts on. The walk is bounded by the cone's own width —
// five hops each way — and needs no subtraction, no wrap, and no sign convention.
func (s *tiltState) acuteWith(target *tiltState) bool {
	if s == target {
		return true
	}
	up, down := s, s
	for hops := int32(1); hops < Wiring.PerpendicularThetaIdx; hops++ {
		up, down = up.next, down.prev
		if up == target || down == target {
			return true
		}
	}
	return false
}
