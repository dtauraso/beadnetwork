package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/rulemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/rulenode"
)

func (m *NodeGeometry) AttachRuleNode(rn *rulenode.RuleNode) {
	m.rule.Attach(rn)
}

func (m *NodeGeometry) RuleNode() *rulenode.RuleNode { return m.rule.Node() }

func (m *NodeGeometry) RuleBackChannel(peerID string) chan rulemsg.Msg {
	return m.rule.Node().RuleBackChannel(peerID)
}

func (m *NodeGeometry) StartRuleNode(ctx context.Context) {
	m.rule.Start(ctx)
}

func (m *NodeGeometry) RuleWake() <-chan struct{} { return m.rule.Wake() }

func (m *NodeGeometry) drainRuleMesh() {
	changed := false
	for {
		state, ok := m.rule.TakeState()
		if !ok {
			break
		}
		m.topo.SetDragRule(state.Rule)
		m.topo.SetDragActive(state.Active)
		m.rule.SetGroup(state.GroupID, state.GroupSize)
		for target, active := range state.EdgeActive {
			m.outEdges.SetEdgeRuleActive(m.id+"To"+target, active)
		}
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
	return m.rule.Group()
}
