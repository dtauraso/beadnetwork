package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
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

// TrimOwnDrag is what a node does to a drag of ITSELF. The drag arrives as the
// polar delta triple it was converted to — the pointer hit is turned into Δ
// against this node's own centre and handed over untouched — and everything
// from here is the node's own answer about how much of that Δ it will take.
//
// It lives on NodeGeometry, in this package, because that is what "the node
// owns its trimming" has to mean if it is to be more than a comment. Every
// input is read off `m`: its own kind, its own orbit rule, its own sides of
// its own edges. Nothing here can reach another node's state, and no other
// package can apply these rules on a node's behalf — `layoutquant` imports
// `nodeactor`, so the trims cannot move back out without a cycle.
//
// It used to run in layoutquant.RootMove, on the pointer's goroutine, before
// the node was told anything. The state it read was already only this node's,
// so the numbers are the same; what changes is that the node is now the thing
// that decides them, and is holding its own rules while it does.
func (m *NodeGeometry) TrimOwnDrag(delta polar.Polar) polar.Polar {
	// An input node's own drag has its theta snapped to the nearest whole
	// multiple of pi the moment it takes the triple, before any rule reads it:
	// it turns half a turn about its own pole or not at all, and nothing
	// downstream ever sees a partial turn to accumulate. Every other node drags
	// with the theta the cursor asked for.
	//
	// This one is genuinely the KIND's, not an id's: it is a statement about
	// what an input node is, and it constrains that node's own drag rather
	// than where anything else may sit. The angles a node holds about the node
	// it hangs from went the other way — they are per-id, carried by the node
	// they bind, and trimToOrbitRule reads them off this node.
	if m.SelfKind() == SharedLengthKind {
		delta = polar.SnapDeltaTheta(delta)
	}
	delta = m.trimToOrbitRule(delta)
	// Its neighbours are not moving, so keeping every outgoing path the same
	// length is a constraint on where THIS node may go.
	return m.trimToEqualOutLengths(delta)
}

// trimToOrbitRule keeps only the part of a drag on this node that its OWN
// orbit rule allows. A node carrying no rule is free and its drag is returned
// untouched — which is every node but the two that state one.
//
// The constraints are held on D — the vector along the edge — because that is
// what they were always about: where a node sits about the one it hangs from,
// not its place in the world. Every one of them is read off this node's own
// side of its own edges, so nothing here reads another node's centre and there
// is no holder frame to convert in and out of.
//
// Nothing here asks what KIND its neighbours are: whether a node orbits is its
// own answer to give, and a holder does not acquire the right to constrain by
// being of some type.
func (m *NodeGeometry) trimToOrbitRule(delta polar.Polar) polar.Polar {
	rule := m.OrbitRule()
	if rule == nil {
		return delta
	}
	for neighborID := range m.NeighborKinds() {
		// An out-target is a node THIS one points at. This node hangs from the
		// other ones, and those are the edges it orbits about.
		if m.IsOutTarget(neighborID) {
			continue
		}
		// The edge runs holder -> here, so its D is this node's own side read
		// from the other end. The drag asks for that same side plus Δ; what
		// the rule allows of it is what the node may take, and the difference
		// between the two is the trimmed drag.
		have, ok := m.DeltaFrom(neighborID)
		if !ok {
			continue
		}
		want := polar.Compose(have, delta)
		delta = polar.Between(have, rule.TrimDelta(have, want))
	}
	return delta
}

// trimToEqualOutLengths keeps every outgoing path of this node the same length,
// when this node is the kind whose paths share one.
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
func (m *NodeGeometry) trimToEqualOutLengths(delta polar.Polar) polar.Polar {
	if m.SelfKind() != SharedLengthKind {
		return delta
	}

	longest, shortest, count := 0.0, 0.0, 0
	for _, to := range m.OutTargets() {
		d, ok := m.DeltaTo(to)
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
