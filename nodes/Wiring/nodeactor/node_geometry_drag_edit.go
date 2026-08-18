package nodeactor

import (
	"github.com/dtauraso/wirefold/tools/topology-vscode/PolarRules"
)

func (m *NodeGeometry) SetDragRuleCopy(rule *PolarRules.DragRule, active bool) {
	m.topo.SetDragRule(rule)
	m.topo.SetDragActive(active)
}
