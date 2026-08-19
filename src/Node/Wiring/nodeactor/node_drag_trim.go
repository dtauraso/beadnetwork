package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodedrag"
	"github.com/dtauraso/wirefold/src/Node/Wiring/polarindex"
)

func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Offset) polarindex.Offset {
	return nodedrag.Apply(m.SelfKind(), delta, m)
}
