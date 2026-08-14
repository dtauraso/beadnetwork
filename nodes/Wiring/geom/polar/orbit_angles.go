package polar

import "math"

// An OrbitRule is a NODE'S OWN statement of how it may sit about the node it
// hangs from. It is carried by the node the rule binds — read off its id, out
// of its own meta.json — not by the holder and not by either one's kind.
//
// That ownership is the whole point. The same numbers used to be two package
// constants (OutAnglePhi, OutAngleMaxTheta) reached through the HOLDER's kind:
// a node was constrained because something of kind "Input" happened to point at
// it, so the rule was stated in a place that could not name which node it was
// about, and every node an Input pointed at got the same one whether or not it
// wanted it.
//
// A node with no rule is FREE — its drag is not trimmed at all. That is why
// adding the rule to two nodes changed nothing for the other seven.
//
// Each field is a component decided ON ITS OWN, the shape every other rule in
// this package has:
//
//   - Phi, if set, is the ONE phi the node may hold. A range would be a second
//     kind of thing; there is no phi range.
//   - MaxTheta, if set, is how far around either way. Inside it theta passes
//     through, outside it stops at the end it crossed.
//   - R is held by the mere EXISTENCE of the rule, in TrimDelta. Holding r is
//     what makes the move an orbit rather than a drift, and it is also what
//     keeps a shared length shared: a length this node cannot change is a
//     length no drag of it can make unequal. It is deliberately not a field —
//     "orbits its holder at a radius it may not state" is what having a rule
//     MEANS, so there is no way to spell an orbit that changes its own radius.
type OrbitRule struct {
	Phi      *float64 `json:"phi,omitempty"`
	MaxTheta *float64 `json:"maxTheta,omitempty"`
}

// ClampPoint puts a path on the allowed angles, whatever it was before, and is
// what runs when a position is ESTABLISHED — at load, where a phi that was
// written wrong has to be corrected rather than merely not made worse.
//
// R always passes through here. At load the length is stated by the holder
// (layoutquant.HeldOutNeighbors), so this must not have an opinion about it.
func (r *OrbitRule) ClampPoint(p Polar) Polar {
	if r == nil {
		return p
	}
	out := p
	if r.Phi != nil {
		out.Phi = *r.Phi
	}
	if r.MaxTheta != nil {
		out.Theta = clampTheta(out.Theta, *r.MaxTheta)
	}
	return out
}

// TrimDelta is the same rule applied to a MOVE rather than to a position:
// `have` is where the path is, `want` where a drag is asking to put it, and the
// result keeps only the part of that delta the rule allows. Whatever is left
// over is DROPPED, not resisted — the drag still moves in every component that
// had room.
//
// It differs from ClampPoint in that the held components keep HAVE's value
// rather than the rule's, so their delta is exactly zero and the drag is absent
// from them rather than performed and undone. It therefore can never CORRECT a
// component that is already wrong, which is why the absolute clamp is the one
// that runs at load.
//
// R is held whenever a rule exists. It used to pass through, and the holder's
// other targets were then teleported onto whatever length this drag had stated
// (HeldSiblings, deleted) — which is what made two nodes hanging off the same
// holder read as welded together.
func (r *OrbitRule) TrimDelta(have, want Polar) Polar {
	if r == nil {
		return want
	}
	out := want
	out.R = have.R
	if r.Phi != nil {
		out.Phi = have.Phi
	}
	if r.MaxTheta != nil {
		out.Theta = clampTheta(out.Theta, *r.MaxTheta)
	}
	return out
}

func clampTheta(theta, max float64) float64 {
	return math.Max(-max, math.Min(max, theta))
}
