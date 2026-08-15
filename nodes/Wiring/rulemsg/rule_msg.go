package rulemsg

import "github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"

type Key struct {
	RLocked bool

	PhiSet bool
	Phi    float64

	ThetaSet bool
	MaxTheta float64
}

func KeyOf(rule *polar.OrbitRule) Key {
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
}
