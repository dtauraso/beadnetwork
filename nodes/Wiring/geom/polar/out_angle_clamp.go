package polar

import "math"

// OutAngleMaxTheta is how far around the pole an outgoing path may turn.
const OutAngleMaxTheta = math.Pi / 2

// ClampOutAngles is the outgoing path an input node is allowed to hold: phi is
// pinned to pi/2 and theta keeps whatever it already is, unless that is past
// +-pi/2, in which case it stops there. R is untouched — an outgoing edge's
// length is what sets the input node's own emission cadence, so the constraint
// is on the two angles alone.
//
// phi = pi/2 is the equator of the node's own pole triple, so the neighbour
// sits at exactly the node's height; |theta| <= pi/2 is the half turn around
// that pole starting at +x.
//
// It takes and returns a Polar, so there is no trigonometry here at all: an
// assignment and a clamp on the two angles. The cartesian boundary is
// somewhere else, which is the only place it should ever be.
func ClampOutAngles(p Polar) Polar {
	p.Phi = math.Pi / 2
	p.Theta = math.Max(-OutAngleMaxTheta, math.Min(OutAngleMaxTheta, p.Theta))
	return p
}
