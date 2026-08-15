package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

func (m *NodeGeometry) SetOrbitRuleCopy(rule *polar.OrbitRule, active bool) {
	m.topo.SetOrbitRule(rule)
	m.topo.SetOrbitActive(active)
}
