package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/rulenode"
)

func (m *NodeGeometry) LinkRuleState(fromRuleNode <-chan rulenode.State, wake <-chan struct{}) {
	m.ruleCopy.LinkRuleNode(fromRuleNode, wake)
}

func (m *NodeGeometry) RuleWake() <-chan struct{} { return m.ruleCopy.Wake() }

func (m *NodeGeometry) drainRuleMesh() {
	changed := false
	for {
		state, ok := m.ruleCopy.TakeState()
		if !ok {
			break
		}
		m.topo.SetOrbitRule(state.Rule)
		m.topo.SetOrbitActive(state.Active)
		m.ruleCopy.SetGroup(state.GroupID, state.GroupSize)
		changed = true
	}
	if !changed {
		return
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

func (m *NodeGeometry) RuleGroup() (groupID, size int32) {
	return m.ruleCopy.Group()
}
