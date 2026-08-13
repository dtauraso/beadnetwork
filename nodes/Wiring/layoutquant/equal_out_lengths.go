package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
)

// TrimEqualOutLengths keeps every outgoing path of a dragged input node the
// same length.
//
// A drag carries ONE r. Every side this node touches loses that same r, so
// paths that were equal stay equal and there is nothing to solve: the rule is
// arithmetic on one component, the way `phi = pi/2` and `|theta| <= pi/2` are
// arithmetic on the other two.
//
// What is left to do is the case where they are NOT already equal — a layout
// that was authored or loaded unequal. Then the drag states one length for all
// of them: the longest, so the short ones grow rather than a node being pulled
// inward past something it was already clear of. That is a change to r and to
// nothing else, so it cannot swing a neighbour sideways.
//
// It used to project the drag onto the plane of points equidistant from the two
// neighbours — dot products, a normal, a nearest point. That treated the equal
// length as a place the node was forbidden to stand, which is what a constraint
// becomes when D is an output of two positions rather than the edge's own
// state. It also removed the whole of any drag along that plane's normal, so
// dragging parallel to the line through two neighbours moved the node 0.000 of
// 30 units.
func TrimEqualOutLengths(nm *nodeactor.NodeGeometry, delta polar.Polar) polar.Polar {
	if nm.SelfKind() != OutAngleKind {
		return delta
	}

	longest, shortest, count := 0.0, 0.0, 0
	for _, to := range nm.OutTargets() {
		d, ok := nm.DeltaTo(to)
		if !ok {
			continue
		}
		if count == 0 || d.R > longest {
			longest = d.R
		}
		if count == 0 || d.R < shortest {
			shortest = d.R
		}
		count++
	}
	if count < 2 || longest == shortest {
		// Already one length. The drag takes the same r off every side, so it
		// stays one length however far this node is dragged.
		return delta
	}

	// Unequal, so this drag states the shared length. Taking less r than asked
	// for is what brings the short sides up to the long one.
	return polar.Polar{R: delta.R - (longest - shortest), Phi: delta.Phi, Theta: delta.Theta}
}
