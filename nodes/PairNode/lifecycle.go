package PairNode

// lifecycle.go — the small per-cycle helpers Update (node.go) calls that are neither the
// composer/builder nor one of the two functions that decide (stepFromVector,
// handleVectorCycle) — node.go's own header names exactly those as what stays there.
// clock() guards a possibly-unset Clock field; openingEmit is everything this node says
// ONCE before its loop starts; paceOnBeadArrival drains In non-blocking to pace (never
// decide) the exchange. Split out of node.go by concern, same-package, no API change.

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

func (n *Node) clock() clock.Clock {
	if n.plumb.Clock == nil {
		return clock.NewRealClock()
	}
	return n.plumb.Clock
}

// openingEmit is everything this node says ONCE, before its loop starts.
//
// Report THIS node's OPENING tilt/normal pair once, before the loop. Self is a
// passive mirror of these (PairNodeSelf.SetTiltIndex) and has no way to derive the
// normal itself, so without this its normal indices sit at their zero value until the
// first arrival or panel click — and since the tilt index opens at 0 too, both
// directions decode to world +y and the two drawn arrows superimpose, which reads as
// the coplanar normal being missing entirely.
func (n *Node) openingEmit() {
	wire.TryEmit(n.plumb.EmitGeometry)
	// This node's own mover-owned startup geometry emit — see Self's own doc comment.
	// There is no separate nodeMover goroutine to make this emit any more.
	n.plumb.Self.EmitGeometryOnce()
	n.syncTiltIndex()
}

// paceOnBeadArrival drains In non-blocking. A bead arrival PACES the exchange and marks the
// round trip; it DECIDES nothing. It used to step this node's tilt one click in this
// kind's own fixed direction, with no reference to anything that arrived — so
// every bead round trip turned this node the same way forever, independently of
// (and on top of) the acute-test rule that is supposed to own that decision. Two
// rules moved one index: when they agreed the node double-stepped, when they
// disagreed they cancelled and it froze. The tests are now the only thing that turns a tilt
// on an arrival, and the bead is what makes that turn visible and timed.
//
// It does not place a bead onward either: the bead now travels WITH the vector,
// placed by handleVectorCycle when the tests actually move this node, so the bead
// loop lives and dies with the exchange it is pacing instead of circulating on
// its own.
func (n *Node) paceOnBeadArrival() {
	if _, ok := n.plumb.In.PollRecv(); ok {
		if n.plumb.Fire != nil {
			n.plumb.Fire()
		}
	}
}
