package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodedrag"
)

func (m *NodeGeometry) TrimOwnDrag(delta polar.Polar) polar.Polar {
	return nodedrag.Apply(m.SelfKind(), delta, m)
}
