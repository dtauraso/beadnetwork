package Node

import (
	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (m *NodeGeometry) SetKindRule(trim NodeDrag.Trim, request NodeDrag.Request) {
	m.kindRule.SetKindRule(trim, request)
}


func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Offset) polarindex.Offset {
	return NodeDrag.Apply(m.kindRule.Trim(), delta, m.dragState())
}

func (m *NodeGeometry) RequestedDrag(delta polarindex.Offset) map[string]polarindex.Offset {
	return NodeDrag.Requested(m.kindRule.Request(), delta, m.dragState())
}

func (m *NodeGeometry) dragState() NodeDrag.State {
	st := NodeDrag.State{
		Index:      m.ComposedIndex(),
		Constants:  m.geom.SceneConstants,
		OutTargets: m.outTargets,
		OutDelta:   make(map[string]polarindex.Offset, len(m.outTargets)),
	}
	for _, to := range m.outTargets {
		if d, ok := m.deltas.DeltaTo(to); ok {
			st.OutDelta[to] = d
		}
	}
	return st
}
