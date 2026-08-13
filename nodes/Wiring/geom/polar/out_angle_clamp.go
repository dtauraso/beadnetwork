package polar

import "math"

// OutAnglePhi is the one phi an outgoing path may hold: the equator of the
// node's own pole triple, which puts the neighbour at exactly the node's
// height.
const OutAnglePhi = math.Pi / 2

// OutAngleMaxTheta is how far around that pole the path may turn either way.
const OutAngleMaxTheta = math.Pi / 2

// ClampOutAngles is the outgoing path an input node is allowed to hold.
//
// Each polar component is decided ON ITS OWN, so a drag that is valid in one
// component still moves in that component even while another is being held:
//
//   - R is unconstrained and always passes through, whatever the angles do.
//     An outgoing edge's length is what sets the input node's emission
//     cadence, so pinning it would silently retime the node.
//   - Phi has exactly one allowed value, so it is always that value. There is
//     no range for a drag to be valid inside.
//   - Theta has a range, so it passes through UNCHANGED whenever it is inside
//     that range, and stops at the end it crossed when it is not. A drag that
//     runs theta past the limit still moves the node in R, and still moves it
//     in theta right up to the limit — it does not freeze the drag.
//
// It takes and returns a Polar, so there is no trigonometry here at all. The
// cartesian boundary is somewhere else, which is the only place it should be.
func ClampOutAngles(p Polar) Polar {
	return Polar{
		R:     p.R,
		Phi:   OutAnglePhi,
		Theta: clampOutTheta(p.Theta),
	}
}

// TrimOutAngleDelta is the same constraint applied to a MOVE rather than to a
// position: `have` is where the path is, `want` is where a drag is asking to
// put it, and the result keeps only the part of that delta the constraint
// allows. Whatever is left over is dropped, not resisted — the drag still
// moves the node in every component that had room.
//
// It differs from ClampOutAngles in exactly one component, and that difference
// is the whole point of having both:
//
//   - R keeps the full delta, as ever.
//   - Phi keeps HAVE's value, not the allowed one. Its delta is always zero,
//     so a drag cannot move phi at all — the forbidden motion is absent from
//     the move rather than performed and then undone. That also means this
//     will never CORRECT a phi that is already wrong, which is why the
//     absolute clamp still runs when a position is established rather than
//     dragged.
//   - Theta keeps as much of the delta as fits and stops at the limit it
//     crossed, so a drag that runs past the edge still slides along it.
func TrimOutAngleDelta(have, want Polar) Polar {
	return Polar{
		R:     want.R,
		Phi:   have.Phi,
		Theta: clampOutTheta(want.Theta),
	}
}

func clampOutTheta(theta float64) float64 {
	return math.Max(-OutAngleMaxTheta, math.Min(OutAngleMaxTheta, theta))
}
