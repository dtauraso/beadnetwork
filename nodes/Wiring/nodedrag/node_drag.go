package nodedrag

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

type Node interface {
	ScenePolar() polar.Polar
	OrbitRule() *polar.OrbitRule
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

func TrimFor(kind string) Trim {
	if t, ok := trims[kind]; ok {
		return t
	}
	return TrimToOrbitRule
}

func RequestFor(kind string) Request {
	if r, ok := requests[kind]; ok {
		return r
	}
	return nil
}

func TrimToOrbitRule(delta polar.Polar, of Node) polar.Polar {
	rule := of.OrbitRule()
	if rule == nil {
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
