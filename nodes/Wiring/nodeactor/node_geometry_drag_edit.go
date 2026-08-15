package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

func (m *NodeGeometry) SetDragRuleCopy(rule *polar.DragRule, active bool) {
	m.topo.SetDragRule(rule)
	m.topo.SetDragActive(active)
}
