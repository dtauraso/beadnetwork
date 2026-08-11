package PairNode

// ring.go — a NODE'S OWN place on the θ lattice: which end a nil field falls back to, which end
// an update names, what survives a change of point count, and the drop test for an arrival on
// the wrong lattice. The lattice itself — the ring of states, the tilt state machine, and the
// pure arithmetic proofs for both — moved to nodes/PairNode/tiltring: it is math with no node, no
// goroutine, no channel, so it sits in its own package with the tests that exercise it.

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/PairNode/tiltring"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

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

// defaultRing is the lattice a node gets when nothing has said otherwise — the count this
// model has always run at, and what a bare test build in this package constructs against. It
// is built once and never written to, so it is immutable shared data rather than shared
// mutable state; a node given a different count builds its own ring instead of touching this.
var defaultRing = tiltring.NewRing(tiltvector.FullTurnThetaIdx)

// A NODE'S OWN PLACE ON THE RING — the accessors every other file reads its lattice and its
// two ends through, and the one function that gives it a different lattice. They live here
// because each is about the ring rather than about the exchange: what a nil field falls back
// to, which end an update names, and what survives a change of point count.

// ringOf is this node's own lattice, with the default standing in for a Ring that was never
// set — a bare test build. Every read of the lattice goes through here.
func (n *Node) ringOf() *tiltring.Ring {
	if n.lattice.Ring == nil {
		return defaultRing
	}
	return n.lattice.Ring
}

// topState is this node's own tilt direction, with its ring's origin standing in for a Top
// that was never set — see the field's own doc comment. Every read of the tilt goes through
// here, so nothing else in this file has to care about that case.
func (n *Node) topState() *tiltring.State {
	if n.tilt.Top == nil {
		return n.ringOf().At(0)
	}
	return n.tilt.Top
}

// bottomState is the other end of the same line, read the same way topState reads the first.
func (n *Node) bottomState() *tiltring.State {
	if n.tilt.Bottom == nil {
		return n.topState().Opposite
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
func (n *Node) setTop(top *tiltring.State) { n.tilt.Top, n.tilt.Bottom = top, top.Opposite }

func (n *Node) setBottom(bottom *tiltring.State) { n.tilt.Bottom, n.tilt.Top = bottom, bottom.Opposite }

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
	return received.Points != 0 && received.Points != n.ringOf().Points
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
	if points == n.ringOf().Points {
		return
	}
	keptIdx := n.topState().Idx
	n.lattice.Ring = tiltring.NewRing(points)
	top, unknown := n.lattice.Ring.SeedState(keptIdx)
	n.setTop(top)
	if unknown && n.plumb.Self != nil {
		n.plumb.Self.Breadcrumb("pair-lattice-adopt", fmt.Sprintf(
			"points=%d keptIdx=%d unknown=true loaded=%d", points, keptIdx, top.Idx))
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
