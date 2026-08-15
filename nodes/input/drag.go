package input

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodedrag"
)

func init() {
	nodedrag.RegisterTrim("Input", trimOwnDrag)
	nodedrag.RegisterRequest("Input", equalOutLengths)
}

func trimOwnDrag(delta polar.Polar, of nodedrag.Node) polar.Polar {
	delta = polar.SnapDeltaTheta(delta)
	delta = nodedrag.TrimToDragRule(delta, of)
	return holdEqualOutLengths(delta, of)
}

func holdEqualOutLengths(delta polar.Polar, of nodedrag.Node) polar.Polar {
	longest, shortest, count := 0.0, 0.0, 0
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
	return polar.Polar{R: delta.R - (longest - shortest), Phi: delta.Phi, Theta: delta.Theta}
}

func equalOutLengths(delta polar.Polar, of nodedrag.Node) map[string]polar.Polar {
	paths := map[string]polar.Polar{}
	shared := 0.0
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
	selfWas := of.ScenePolar()
	selfNow := polar.Compose(selfWas, delta)
	out := make(map[string]polar.Polar, len(paths))
	for to, p := range paths {
		wants := p
		wants.R = shared
		out[to] = polar.Between(polar.Compose(selfWas, p), polar.Compose(selfNow, wants))
	}
	return out
}
