package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
)

func HeldOutNeighbors(
	nm *nodeactor.NodeGeometry,
	delta polar.Polar,
	ruleOf func(id string) *polar.OrbitRule,
) map[string]polar.Polar {
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
	selfPoint := polar.Compose(nm.ScenePolar(), delta)
	out := make(map[string]polar.Polar, len(paths))
	for to, p := range paths {
		held := ruleOf(to).ClampPoint(p)
		held.R = shared
		out[to] = polar.Compose(selfPoint, held)
	}
	return out
}
