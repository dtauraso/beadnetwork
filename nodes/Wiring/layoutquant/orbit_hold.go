package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
)

// SharedLengthKind is the one kind whose outgoing paths share a length. It is
// the SPEC kind name, which is PascalCase — the Go package directory for this
// kind is lowercase `nodes/input/` and the two do not have to agree.
//
// It is the ONLY thing left that names a kind. It used to gate the angles too,
// under the name OutAngleKind, and that was the bug: "phi = pi/2, |theta| <=
// pi/2" was stated as two package constants reached through the HOLDER's kind,
// so the rule had no way to name the node it was actually about and every node
// an Input pointed at inherited the same one. The angles now live on the node
// they bind, by id, in its own meta.json (polar.OrbitRule).
//
// What genuinely remains this kind's, because it is about what an input node
// IS rather than about where any particular node sits: the shared length
// across its outgoing paths (which is what sets its emission cadence, so the
// paths must fire on one beat) and the half-turn snap on its own drag.
const SharedLengthKind = "Input"

// The constraints are enforced HERE, where a drag is turned into positions, and
// nowhere else. No node holds a path to another node, nothing is cached from a
// broadcast, and no node moves another in response to a message — a drag
// decides what moves, which is the same thing group-length editing already
// does through this entry point.
//
// They are held on D — the vector along the edge — because that is what they
// were always about: where a node sits about the one it hangs from, not its
// place in the world. Every one of them is read off the node's own side of its
// own edges, so nothing here reads another node's centre and there is no holder
// frame to convert in and out of.

// TrimDraggedNode keeps only the part of a drag on this node that its OWN
// orbit rule allows. A node carrying no rule is free and its drag is returned
// untouched — which is every node but the two that state one.
//
// The rule is read off `nm`, the node under the cursor. Nothing here asks what
// KIND its neighbours are: whether a node orbits is its own answer to give, and
// a holder does not acquire the right to constrain by being of some type.
func TrimDraggedNode(nm *nodeactor.NodeGeometry, delta polar.Polar) polar.Polar {
	rule := nm.OrbitRule()
	if rule == nil {
		return delta
	}
	for neighborID := range nm.NeighborKinds() {
		// An out-target is a node THIS one points at. This node hangs from the
		// other ones, and those are the edges it orbits about.
		if nm.IsOutTarget(neighborID) {
			continue
		}
		// The edge runs holder -> here, so its D is this node's own side read
		// from the other end. The drag asks for that same side plus Δ; what
		// the rule allows of it is what the node may take, and the difference
		// between the two is the trimmed drag.
		have, ok := nm.DeltaFrom(neighborID)
		if !ok {
			continue
		}
		want := polar.Compose(have, delta)
		delta = polar.Between(have, rule.TrimDelta(have, want))
	}
	return delta
}

// HeldSiblings is DELETED, and nothing replaced it.
//
// It moved node 3 whenever node 2 was dragged. Its premise was stated in its
// own comment: "moving node 2 changes |1->2|, and nothing about node 2's own
// position can bring |1->3| along with it", so the dragged node stated a new
// shared length and its siblings were teleported onto it. That premise is now
// false. TrimDraggedNode holds R, so a drag of node 2 does not change |1->2|
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
	if nm.SelfKind() != SharedLengthKind {
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
