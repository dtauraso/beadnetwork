package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/spatial"
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
// Centres come from the registry's mirror of each node's OWN published
// centre. That is a value the node sent, not a field read out from under it.

// TrimDraggedNode keeps only the part of a drag on `nodeID` that its own
// constraints allow. A node with no input node pointing at it is unaffected.
func TrimDraggedNode(nm *nodeactor.NodeGeometry, centerOf func(string) (spatial.Vec3, bool), from, target spatial.Vec3) spatial.Vec3 {
	for neighborID, kind := range nm.NeighborKinds() {
		// An out-target is a node THIS one points at, whose angles are that
		// node's business, not a constraint on where this one may sit.
		if kind != OutAngleKind || nm.IsOutTarget(neighborID) {
			continue
		}
		center, ok := centerOf(neighborID)
		if !ok {
			continue
		}
		have := polar.Cart2polar(from.Sub(center))
		want := polar.Cart2polar(target.Sub(center))
		target = center.Add(polar.Polar2cart(polar.TrimOutAngleDelta(have, want)))
	}
	return target
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
	centerOf func(string) (spatial.Vec3, bool),
	draggedID string,
	target spatial.Vec3,
) map[string]spatial.Vec3 {
	var held map[string]spatial.Vec3
	for holderID, kind := range nm.NeighborKinds() {
		if kind != OutAngleKind || nm.IsOutTarget(holderID) {
			continue
		}
		holder, ok := nodeGeoms[holderID]
		if !ok {
			continue
		}
		holderCenter, ok := centerOf(holderID)
		if !ok {
			continue
		}
		shared := target.Sub(holderCenter).Length()
		for _, sib := range holder.OutTargets() {
			if sib == draggedID {
				continue
			}
			c, ok := centerOf(sib)
			if !ok {
				continue
			}
			p := polar.ClampOutAngles(polar.Cart2polar(c.Sub(holderCenter)))
			p.R = shared
			if held == nil {
				held = map[string]spatial.Vec3{}
			}
			held[sib] = holderCenter.Add(polar.Polar2cart(p))
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
func HeldOutNeighbors(nm *nodeactor.NodeGeometry, centerOf func(string) (spatial.Vec3, bool), selfTarget spatial.Vec3) map[string]spatial.Vec3 {
	if nm.SelfKind() != OutAngleKind {
		return nil
	}
	paths := map[string]polar.Polar{}
	shared := 0.0
	for _, to := range nm.OutTargets() {
		c, ok := centerOf(to)
		if !ok {
			continue
		}
		p := polar.Cart2polar(c.Sub(selfTarget))
		paths[to] = p
		if p.R > shared {
			shared = p.R
		}
	}
	if len(paths) == 0 {
		return nil
	}
	out := make(map[string]spatial.Vec3, len(paths))
	for to, p := range paths {
		held := polar.ClampOutAngles(p)
		held.R = shared
		out[to] = selfTarget.Add(polar.Polar2cart(held))
	}
	return out
}
