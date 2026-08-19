package nodeactor

import (
	"github.com/dtauraso/wirefold/tools/topology-vscode/src/PolarRulesPanel"
)

func (m *NodeGeometry) SetDragRuleCopy(rule *PolarRulesPanel.DragRule, active bool) {
	m.topo.SetDragRule(rule)
	m.topo.SetDragActive(active)
}
