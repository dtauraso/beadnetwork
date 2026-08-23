package nodeactor

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodedrag"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (m *NodeGeometry) SetKindRule(trim nodedrag.Trim, request nodedrag.Request) {
	m.kindRule.SetKindRule(trim, request)
}

func (m *NodeGeometry) HasKindRule() bool { return m.kindRule.HasKindRule() }

func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Offset) polarindex.Offset {
	return nodedrag.Apply(m.kindRule.Trim(), delta, m)
}

func (m *NodeGeometry) RequestedDrag(delta polarindex.Offset) map[string]polarindex.Offset {
	return nodedrag.Requested(m.kindRule.Request(), delta, m)
}
