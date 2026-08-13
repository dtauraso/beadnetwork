package polar

import "math"

// OutAnglePhi is the one phi an input node's outgoing path may hold: the
// equator of that node's own pole triple, which puts the neighbour at exactly
// the node's height.
const OutAnglePhi = math.Pi / 2

// OutAngleMaxTheta is how far around that pole the path may turn either way.
const OutAngleMaxTheta = math.Pi / 2

// ClampOutAngles puts a path on the allowed angles, whatever it was before.
// Each polar component is decided ON ITS OWN:
//
//   - R is unconstrained and always passes through. An outgoing edge's length
//     is what sets the input node's emission cadence, so pinning it would
//     silently retime the node.
//   - Phi has exactly one allowed value, so it is always that value.
//   - Theta has a range, so it passes through unchanged inside that range and
//     stops at the end it crossed outside it.
func ClampOutAngles(p Polar) Polar {
	return Polar{
		R:     p.R,
		Phi:   OutAnglePhi,
		Theta: clampOutTheta(p.Theta),
	}
}

// TrimOutAngleDelta is the same constraint applied to a MOVE rather than to a
// position: `have` is where the path is, `want` where a drag is asking to put
// it, and the result keeps only the part of that delta the constraint allows.
// Whatever is left over is dropped, not resisted — the drag still moves in
// every component that had room.
//
// It differs from ClampOutAngles in exactly one component: phi keeps HAVE's
// value, so its delta is always zero and a phi drag is absent from the move
// rather than performed and undone. That also means it can never CORRECT a
// phi that is already wrong, which is why the absolute clamp is what runs
// when a position is established rather than dragged.
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
