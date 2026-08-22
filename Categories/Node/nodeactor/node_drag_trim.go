package nodeactor

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodedrag"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Offset) polarindex.Offset {
	return nodedrag.Apply(m.SelfKind(), delta, m)
}
