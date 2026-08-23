package Node

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodedrag"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (m *NodeGeometry) SetKindRule(trim nodedrag.Trim, request nodedrag.Request) {
	m.kindRule.SetKindRule(trim, request)
}

func (m *NodeGeometry) HasKindRule() bool { return m.kindRule.HasKindRule() }

func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Offset) polarindex.Offset {
	return nodedrag.Apply(m.kindRule.Trim(), delta, m.dragState())
}

func (m *NodeGeometry) RequestedDrag(delta polarindex.Offset) map[string]polarindex.Offset {
	return nodedrag.Requested(m.kindRule.Request(), delta, m.dragState())
}

func (m *NodeGeometry) dragState() nodedrag.State {
	st := nodedrag.State{
		Index:      m.ComposedIndex(),
		Constants:  m.geom.SceneConstants,
		Drag:       m.topo.DragRule(),
		DragOn:     m.topo.DragRuleActive(),
		Self:       m.topo.SelfRule(),
		SelfOn:     m.topo.SelfRuleActive(),
		KindOn:     m.KindRuleActive(),
		OutTargets: m.outTargets,
		OutDelta:   make(map[string]polarindex.Offset, len(m.outTargets)),
		Inbound:    map[string]polarindex.Offset{},
	}
	for _, to := range m.outTargets {
		if d, ok := m.deltas.DeltaTo(to); ok {
			st.OutDelta[to] = d
		}
	}
	for neighborID := range m.topo.NeighborKinds() {
		if m.IsOutTarget(neighborID) || !m.EdgeRuleActive(neighborID) {
			continue
		}
		if d, ok := m.deltas.DeltaFrom(neighborID); ok {
			st.Inbound[neighborID] = d
		}
	}
	return st
}
