package nodedrag

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

type Node interface {
	ScenePolar() polar.Polar
	OrbitRule() *polar.OrbitRule
	OrbitActive() bool
	NeighborKinds() map[string]string
	IsOutTarget(neighborID string) bool
	DeltaFrom(otherID string) (polar.Polar, bool)
	OutTargets() []string
	DeltaTo(otherID string) (polar.Polar, bool)
}

type Trim func(delta polar.Polar, of Node) polar.Polar

type Request func(delta polar.Polar, of Node) map[string]polar.Polar

var trims = map[string]Trim{}

var requests = map[string]Request{}

func RegisterTrim(kind string, t Trim) {
	if _, exists := trims[kind]; exists {
		panic("nodedrag.RegisterTrim: kind already has a trim: " + kind)
	}
	trims[kind] = t
}

func RegisterRequest(kind string, r Request) {
	if _, exists := requests[kind]; exists {
		panic("nodedrag.RegisterRequest: kind already has a request: " + kind)
	}
	requests[kind] = r
}

func HasKindRule(kind string) bool {
	if _, ok := trims[kind]; ok {
		return true
	}
	_, ok := requests[kind]
	return ok
}

func Apply(kind string, delta polar.Polar, of Node) polar.Polar {
	if !of.OrbitActive() {
		return delta
	}
	if t, ok := trims[kind]; ok {
		return t(delta, of)
	}
	return TrimToOrbitRule(delta, of)
}

func Requested(kind string, delta polar.Polar, of Node) map[string]polar.Polar {
	if !of.OrbitActive() {
		return nil
	}
	r, ok := requests[kind]
	if !ok {
		return nil
	}
	return r(delta, of)
}

func TrimToOrbitRule(delta polar.Polar, of Node) polar.Polar {
	rule := of.OrbitRule()
	if rule == nil || !of.OrbitActive() {
		return delta
	}
	for neighborID := range of.NeighborKinds() {
		if of.IsOutTarget(neighborID) {
			continue
		}
		have, ok := of.DeltaFrom(neighborID)
		if !ok {
			continue
		}
		want := polar.Compose(have, delta)
		delta = polar.Between(have, rule.TrimDelta(have, want))
	}
	return delta
}
