package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/rulemsg"
)

func (m *NodeGeometry) RuleBackChannel(peerID string) chan rulemsg.Msg {
	return m.ruleMesh.RuleBackChannel(peerID)
}

func (m *NodeGeometry) LinkRuleDown(peerID string, down chan rulemsg.Msg) {
	m.ruleMesh.LinkRuleDown(peerID, down)
}

func (m *NodeGeometry) BroadcastRule() {
	m.ruleMesh.SetSelfRuleKey(rulemsg.KeyOf(m.OrbitRule()))
	m.ruleMesh.BroadcastRule(m.id)
}

func (m *NodeGeometry) drainRuleMesh() {
	if !m.ruleMesh.DrainRules() {
		return
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

func (m *NodeGeometry) RuleGroup() (groupID, size int32) {
	return m.ruleMesh.RuleGroup(m.id)
}
