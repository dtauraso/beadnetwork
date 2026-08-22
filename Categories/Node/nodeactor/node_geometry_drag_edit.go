package nodeactor

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
)

func (m *NodeGeometry) SetDragRuleCopy(rule *PolarRulesPanel.DragRule, active bool) {
	m.topo.SetDragRule(rule)
	m.topo.SetDragActive(active)
}
