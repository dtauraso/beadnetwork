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
	DeltaFrom(otherID string) (polarindex.Offset, bool)
	OutTargets() []string
	DeltaTo(otherID string) (polarindex.Offset, bool)
}

type Trim func(delta polarindex.Offset, of Node) polarindex.Offset

type Request func(delta polarindex.Offset, of Node) map[string]polarindex.Offset

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

func Apply(kind string, delta polarindex.Offset, of Node) polarindex.Offset {
	if !of.DragRuleActive() {
		return delta
	}
	if t, ok := trims[kind]; ok {
		return t(delta, of)
	}
	return TrimToDragRule(delta, of)
}

func Requested(kind string, delta polarindex.Offset, of Node) map[string]polarindex.Offset {
	if !of.DragRuleActive() {
		return nil
	}
	r, ok := requests[kind]
	if !ok {
		return nil
	}
	return r(delta, of)
}

func TrimToDragRule(delta polarindex.Offset, of Node) polarindex.Offset {
	rule := of.DragRule()
	if rule == nil || !of.DragRuleActive() {
		return delta
	}
	sc := of.Constants()
	for neighborID := range of.NeighborKinds() {
		if of.IsOutTarget(neighborID) {
			continue
		}
		haveOff, ok := of.DeltaFrom(neighborID)
		if !ok {
			continue
		}
		have := polarindex.OffsetToPolar(haveOff, sc)
		wantOff := polarindex.Sum(haveOff, delta)
		want := polarindex.OffsetToPolar(wantOff, sc)
		trimmed := rule.TrimDelta(have, want)
		trimmedOff := polarindex.MeasureOffset(trimmed, sc)
		delta = polarindex.Sum(trimmedOff, polarindex.Neg(haveOff))
	}
	return delta
}
