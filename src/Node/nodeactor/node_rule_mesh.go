package nodeactor

import (
	"context"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/owners"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"

	"github.com/dtauraso/wirefold/src/Node/rulenode"
)

func (m *NodeGeometry) AttachRuleNode(rn *rulenode.RuleNode) {
	m.rule.Attach(rn)
}

func (m *NodeGeometry) RuleNode() *rulenode.RuleNode { return m.rule.Node() }

func (m *NodeGeometry) RuleBackChannel(peerID string) chan PolarRulesPanel.Msg {
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
		m.topo.SetSelfRule(state.SelfRule)
		m.topo.SetSelfRuleActive(state.SelfActive)
		m.rule.SetGroup(state.GroupID, state.GroupSize)
		peerCenters := make(map[string]owners.Vec3, len(state.PeerCenters))
		for id, c := range state.PeerCenters {
			peerCenters[id] = owners.Vec3(c)
		}
		m.channels.SetPeerCenters(peerCenters)
		for target, active := range state.EdgeActive {
			m.outEdges.SetEdgeRuleActive(m.id+"To"+target, active)
		}
		changed = true
	}
	if !changed {
		return
	}
	m.emitGeometry()
}

func (m *NodeGeometry) RuleGroup() (groupID, size int32) {
	return m.rule.Group()
}
