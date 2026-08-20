package PolarRulesPanel

import (
	"math"

	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/spatial"
)

type DragRule struct {
	R        *float64
	Phi      *float64
	MaxTheta *float64
}

func (r *DragRule) TrimDelta(have, want polar.Polar) polar.Polar {
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

type Key struct {
	RLocked bool

	PhiSet bool
	Phi    float64

	ThetaSet bool
	MaxTheta float64
}

func KeyOf(rule *DragRule) Key {
	if rule == nil {
		return Key{}
	}
	k := Key{RLocked: true}
	if rule.Phi != nil {
		k.PhiSet, k.Phi = true, *rule.Phi
	}
	if rule.MaxTheta != nil {
		k.ThetaSet, k.MaxTheta = true, *rule.MaxTheta
	}
	return k
}

type Msg struct {
	FromID string
	Key    Key

	Center    spatial.Vec3
	HasCenter bool
}
