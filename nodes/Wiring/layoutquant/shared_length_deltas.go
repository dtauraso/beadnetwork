package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
)

func SharedLengthDeltas(nm *nodeactor.NodeGeometry, delta polar.Polar) map[string]polar.Polar {
	if nm.SelfKind() != nodeactor.SharedLengthKind {
		return nil
	}
	paths := map[string]polar.Polar{}
	shared := 0.0
	for _, to := range nm.OutTargets() {
		p, ok := nm.DeltaTo(to)
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
	selfWas := nm.ScenePolar()
	selfNow := polar.Compose(selfWas, delta)
	out := make(map[string]polar.Polar, len(paths))
	for to, p := range paths {
		wants := p
		wants.R = shared
		out[to] = polar.Between(polar.Compose(selfWas, p), polar.Compose(selfNow, wants))
	}
	return out
}
