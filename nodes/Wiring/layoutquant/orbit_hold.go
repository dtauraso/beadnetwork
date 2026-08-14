package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
)

// SharedLengthKind, and the trims that read it, MOVED TO THE NODE.
//
// They were package functions here, taking a *nodeactor.NodeGeometry and
// reaching into it: a drag was converted, snapped, trimmed against the node's
// own orbit rule and its own equal out-lengths, and only then was the node told
// where it had ended up. Every number came from the node and none of the
// decisions did. They are now nodeactor.NodeGeometry.TrimOwnDrag and its two
// unexported halves, run on the node's own goroutine off the raw Δ it is sent
// (nodes/Wiring/nodeactor/node_drag_trim.go), and the kind name went with them
// as nodeactor.SharedLengthKind.
//
// Do not restate any of them here. This package imports nodeactor, so a rule
// written on this side can reach into a node's state and a rule written on that
// side cannot reach anyone else's — which is why the trims are over there
// rather than merely described as the node's.

// HeldSiblings is DELETED, and nothing replaced it.
//
// It moved node 3 whenever node 2 was dragged. Its premise was stated in its
// own comment: "moving node 2 changes |1->2|, and nothing about node 2's own
// position can bring |1->3| along with it", so the dragged node stated a new
// shared length and its siblings were teleported onto it. That premise is now
// false. The node's own orbit trim holds R, so a drag of node 2 does not change |1->2|
// at all — there is no broken length for a sibling to be brought to.
//
// What it cost while it existed: 2 and 3 read as welded together and neither
// could be positioned on its own, because every drag of one restated the
// radius of the other. Do not bring it back to "keep the lengths equal" —
// they are equal because no drag of a target can make them unequal, which is
// a stronger statement than a repair that runs afterwards.

// HeldOutNeighbors is where an input node's outgoing neighbours have to be,
// given where that node is going. It answers for the case the dragged node
// cannot: a neighbour does not move itself to satisfy someone else's
// constraint, so the drag that breaks it is the drag that repairs it.
//
// The shared length is the LONGEST current path. Growing the short one is the
// choice that never pulls a node inward past something it was already clear
// of.
//
// The two halves of what it holds now come from two different owners, and that
// is deliberate. The LENGTH is this node's — it is what sets its emission
// cadence, so it is stated once for every outgoing path. The ANGLES are each
// TARGET's, read through `ruleOf` off that target's own id; this node supplies
// none of them and does not know what they are. A target with no rule keeps the
// angles it loaded with and only takes the shared length.
//
// It runs at LOAD with a zero delta, which is the only occasion an absolute
// clamp is right: a phi written wrong on disk cannot be corrected by any move
// of the node it is wrong for, because a move only ever trims a delta.
//
// Reading each target's rule from here is safe for the same reason reading
// their out-targets and kinds is: it is set once at build and never written
// again.
func HeldOutNeighbors(
	nm *nodeactor.NodeGeometry,
	delta polar.Polar,
	ruleOf func(id string) *polar.OrbitRule,
) map[string]polar.Polar {
	if nm.SelfKind() != nodeactor.SharedLengthKind {
		return nil
	}
	// The neighbours COME ALONG. Dragging this node moves it and its
	// out-neighbours by the same Δ, which leaves every D untouched — so
	// phi, theta and the shared length all still hold, exactly, with nothing
	// to re-solve.
	//
	// They used to be treated as standing still, each path recomputed as D
	// composed with -Δ. A drag parallel to the line through two neighbours
	// lengthens one path and shortens the other; the shared length is then the
	// LONGER of the two, so the short one was stretched and its neighbour shoved
	// along its own direction into whatever was there. Moving node 1 twenty
	// units along the 2->3 direction pulled node 3 from r=20.03 to 14.92 and put
	// it nearer node 5 than node 2, while node 2 did not follow at all.
	paths := map[string]polar.Polar{}
	shared := 0.0
	for _, to := range nm.OutTargets() {
		p, ok := nm.DeltaTo(to)
		if !ok {
			continue
		}
		paths[to] = p
		if p.R > shared {
			shared = p.R
		}
	}
	if len(paths) == 0 {
		return nil
	}
	selfPoint := polar.Compose(nm.ScenePolar(), delta)
	out := make(map[string]polar.Polar, len(paths))
	for to, p := range paths {
		held := ruleOf(to).ClampPoint(p)
		held.R = shared
		out[to] = polar.Compose(selfPoint, held)
	}
	return out
}
