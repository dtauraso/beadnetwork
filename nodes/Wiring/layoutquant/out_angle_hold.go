package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
)

// OutAngleKind is the one kind whose outgoing paths are constrained. It is
// the SPEC kind name, which is PascalCase — the Go package directory for this
// kind is lowercase `nodes/input/` and the two do not have to agree.
const OutAngleKind = "Input"

// The constraints an input node's outgoing paths hold, from that node's own
// pole triple: phi = pi/2, |theta| <= pi/2, and ONE shared length across all
// of them, so 1->2 and 1->3 stay the same distance whichever end moved.
//
// They are enforced HERE, where a drag is turned into positions, and nowhere
// else. No node holds a path to another node, nothing is cached from a
// broadcast, and no node moves another in response to a message — a drag
// decides what moves, which is the same thing group-length editing already
// does through this entry point.
//
// The constraints are held on D — the vector along the edge — because that is
// what they were always about: an edge's direction out of its source, not a
// node's place in the world. Every one of them is now read off the node's own
// side of its own edges, so nothing here reads another node's centre and there
// is no holder frame to convert in and out of.

// TrimDraggedNode keeps only the part of a drag on `nodeID` that its own
// constraints allow. A node with no input node pointing at it is unaffected.
func TrimDraggedNode(nm *nodeactor.NodeGeometry, delta polar.Polar) polar.Polar {
	for neighborID, kind := range nm.NeighborKinds() {
		// An out-target is a node THIS one points at, whose angles are that
		// node's business, not a constraint on where this one may sit.
		if kind != OutAngleKind || nm.IsOutTarget(neighborID) {
			continue
		}
		// The edge runs holder -> here, so its D is this node's own side read
		// from the other end. The drag asks for that same side plus Δ; what
		// the constraint allows of it is what the node may take, and the
		// difference between the two is the trimmed drag.
		have, ok := nm.DeltaFrom(neighborID)
		if !ok {
			continue
		}
		want := polar.Compose(have, delta)
		delta = polar.Between(have, polar.TrimOutAngleDelta(have, want))
	}
	return delta
}

// HeldSiblings is where the OTHER outgoing neighbours of an input node have
// to be once one of them has been dragged.
//
// The shared length is the one constraint a dragged node cannot satisfy by
// itself: moving node 2 changes |1->2|, and nothing about node 2's own
// position can bring |1->3| along with it. The node that was dragged states
// the new length — its drag is kept in full — and its siblings are the ones
// brought to it, so the node under the cursor is never the one corrected.
//
// Reading another node's out-targets and kinds is safe: both are set once at
// build and never written again.
func HeldSiblings(
	nm *nodeactor.NodeGeometry,
	nodeGeoms map[string]*nodeactor.NodeGeometry,
	draggedID string,
	delta polar.Polar,
) map[string]polar.Polar {
	var held map[string]polar.Polar
	for holderID, kind := range nm.NeighborKinds() {
		if kind != OutAngleKind || nm.IsOutTarget(holderID) {
			continue
		}
		holder, ok := nodeGeoms[holderID]
		if !ok {
			continue
		}
		// The holder's own side of the edge that reaches this node locates
		// the holder — its point is this node's point less that side — and
		// the drag restates that side's r, which is the r every one of its
		// siblings has to take.
		toHere, ok := nm.DeltaFrom(holderID)
		if !ok {
			continue
		}
		holderPoint := polar.Compose(nm.ScenePolar(), toHere.Neg())
		shared := polar.Compose(toHere, delta).R
		for _, sib := range holder.OutTargets() {
			if sib == draggedID {
				continue
			}
			toSib, ok := holder.DeltaTo(sib)
			if !ok {
				continue
			}
			p := polar.ClampOutAngles(toSib)
			p.R = shared
			if held == nil {
				held = map[string]polar.Polar{}
			}
			held[sib] = polar.Compose(holderPoint, p)
		}
	}
	return held
}

// HeldOutNeighbors is where an input node's outgoing neighbours have to be,
// given where that node is going. It answers for the case the dragged node
// cannot: a neighbour does not move itself to satisfy someone else's
// constraint, so the drag that breaks it is the drag that repairs it.
//
// The shared length is the LONGEST current path. Growing the short one is the
// choice that never pulls a node inward past something it was already clear
// of.
func HeldOutNeighbors(nm *nodeactor.NodeGeometry, delta polar.Polar) map[string]polar.Polar {
	if nm.SelfKind() != OutAngleKind {
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
		held := polar.ClampOutAngles(p)
		held.R = shared
		out[to] = polar.Compose(selfPoint, held)
	}
	return out
}
