package nodedrag

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

type Node interface {
	ScenePolar() polar.Polar
	ComposedIndex() polarindex.Index
	Constants() polarindex.SceneConstants
	DragRule() *polar.DragRule
	DragRuleActive() bool
	NeighborKinds() map[string]string
	IsOutTarget(neighborID string) bool
	DeltaFrom(otherID string) (polarindex.Index, bool)
	OutTargets() []string
	DeltaTo(otherID string) (polarindex.Index, bool)
}

type Trim func(delta polarindex.Index, of Node) polarindex.Index

type Request func(delta polarindex.Index, of Node) map[string]polarindex.Index

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

func Apply(kind string, delta polarindex.Index, of Node) polarindex.Index {
	if !of.DragRuleActive() {
		return delta
	}
	if t, ok := trims[kind]; ok {
		return t(delta, of)
	}
	return TrimToDragRule(delta, of)
}

func Requested(kind string, delta polarindex.Index, of Node) map[string]polarindex.Index {
	if !of.DragRuleActive() {
		return nil
	}
	r, ok := requests[kind]
	if !ok {
		return nil
	}
	return r(delta, of)
}

func TrimToDragRule(delta polarindex.Index, of Node) polarindex.Index {
	rule := of.DragRule()
	if rule == nil || !of.DragRuleActive() {
		return delta
	}
	sc := of.Constants()
	for neighborID := range of.NeighborKinds() {
		if of.IsOutTarget(neighborID) {
			continue
		}
		haveIdx, ok := of.DeltaFrom(neighborID)
		if !ok {
			continue
		}
		have := polarindex.ToPolar(haveIdx, sc)
		wantIdx := polarindex.Compose(haveIdx, delta, sc)
		want := polarindex.ToPolar(wantIdx, sc)
		trimmed := rule.TrimDelta(have, want)
		trimmedIdx := polarindex.Canonical(polarindex.MeasureScalar(trimmed, sc), sc)
		delta = polarindex.Delta(trimmedIdx, haveIdx)
	}
	return delta
}
