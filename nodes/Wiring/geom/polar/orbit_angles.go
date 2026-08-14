package polar

import "math"

type OrbitRule struct {
	Phi      *float64 `json:"phi,omitempty"`
	MaxTheta *float64 `json:"maxTheta,omitempty"`
}

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
