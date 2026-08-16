package polar

import "math"

type DragRule struct {
	R        *float64 `json:"r,omitempty"`
	Phi      *float64 `json:"phi,omitempty"`
	MaxTheta *float64 `json:"maxTheta,omitempty"`
}

func (r *DragRule) TrimDelta(have, want Polar) Polar {
	if r == nil {
		return want
	}
	out := want
	if r.R != nil {
		out.R = have.R
	}
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
