package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Node/nodedrag"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
)

func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Offset) polarindex.Offset {
	return nodedrag.Apply(m.SelfKind(), delta, m)
}
