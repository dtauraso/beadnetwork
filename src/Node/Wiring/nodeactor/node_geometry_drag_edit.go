package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Chrome/PolarRulesPanel"
)

func (m *NodeGeometry) SetDragRuleCopy(rule *PolarRulesPanel.DragRule, active bool) {
	m.topo.SetDragRule(rule)
	m.topo.SetDragActive(active)
}
