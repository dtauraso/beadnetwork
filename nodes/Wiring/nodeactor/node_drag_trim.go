package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/nodedrag"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Index) polarindex.Index {
	return nodedrag.Apply(m.SelfKind(), delta, m)
}
