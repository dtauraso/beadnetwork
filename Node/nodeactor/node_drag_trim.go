package nodeactor

import (
	"github.com/dtauraso/wirefold/Node/nodedrag"
	"github.com/dtauraso/wirefold/Polar/polarindex"
)

func (m *NodeGeometry) TrimOwnDrag(delta polarindex.Offset) polarindex.Offset {
	return nodedrag.Apply(m.SelfKind(), delta, m)
}
