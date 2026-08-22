package nodeactor

import (
	"github.com/dtauraso/wirefold/Node/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/Scene/loadspec"
)

func (m *NodeGeometry) SeedFromSpec(n loadspec.Node, sceneRoot string) {
	m.SetSelfKind(n.Type)

	rn := m.RuleNode()
	rn.SetPersistRoot(sceneRoot)

	active := nodefiles.LoadDragActive(sceneRoot, n.ID)
	m.SetDragRule(n.Drag)
	m.SetDragActive(active)
	rn.SeedRule(n.Drag, active)
	rn.SeedKindActive(nodefiles.LoadKindRuleActive(sceneRoot, n.ID))

	selfActive := nodefiles.LoadSelfRuleActive(sceneRoot, n.ID)
	m.SetSelfRule(n.SelfDrag)
	m.SetSelfRuleActive(selfActive)
	rn.SeedSelfRule(n.SelfDrag, selfActive)

	if n.TopTiltVectorPhiIdx != nil {
		m.SetTopTiltVectorPhiIdx(*n.TopTiltVectorPhiIdx)
	}
}
