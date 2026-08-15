package input

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodedrag"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

func init() {
	nodedrag.RegisterTrim("Input", trimOwnDrag)
	nodedrag.RegisterRequest("Input", equalOutLengths)
}

func trimOwnDrag(delta polarindex.Offset, of nodedrag.Node) polarindex.Offset {
	if !of.KindRuleActive() {
		return nodedrag.TrimToDragRule(delta, of)
	}
	delta = snapDeltaTheta(delta, of.Constants())
	delta = nodedrag.TrimToDragRule(delta, of)
	return holdEqualOutLengths(delta, of)
}

func snapDeltaTheta(delta polarindex.Offset, sc polarindex.SceneConstants) polarindex.Offset {
	if sc.MaxIndexTheta == 0 {
		return delta
	}
	half := sc.MaxIndexTheta / 2
	if half == 0 {
		return delta
	}
	steps := int(math.Round(float64(delta.Theta) / float64(half)))
	delta.Theta = steps * half
	return delta
}

func holdEqualOutLengths(delta polarindex.Offset, of nodedrag.Node) polarindex.Offset {
	longest, shortest, count := 0, 0, 0
	for _, to := range of.OutTargets() {
		d, ok := of.DeltaTo(to)
		if !ok {
			continue
		}
		if count == 0 || d.R > longest {
			longest = d.R
		}
		if count == 0 || d.R < shortest {
			shortest = d.R
		}
		count++
	}
	if count < 2 || longest == shortest {
		return delta
	}
	delta.R -= longest - shortest
	return delta
}

func equalOutLengths(delta polarindex.Offset, of nodedrag.Node) map[string]polarindex.Offset {
	sc := of.Constants()
	paths := map[string]polarindex.Offset{}
	shared := 0
	for _, to := range of.OutTargets() {
		p, ok := of.DeltaTo(to)
		if !ok {
			continue
		}
		paths[to] = p
		if p.R > shared {
			shared = p.R
		}
	}
	if len(paths) == 0 {
		return nil
	}
	selfWas := of.ComposedIndex()
	selfNow := polarindex.Compose(selfWas, delta, sc)
	out := make(map[string]polarindex.Offset, len(paths))
	for to, p := range paths {
		wants := p
		wants.R = shared
		targetOld := polarindex.Compose(selfWas, p, sc)
		targetNew := polarindex.Compose(selfNow, wants, sc)
		out[to] = polarindex.Delta(targetNew, targetOld)
	}
	return out
}
