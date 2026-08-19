package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/polarindex"
)

func (m *NodeGeometry) ApplyCenter(idx polarindex.Index) {
	nodegeom.SetNodeWorld(&m.geom, idx)

	m.msg.PublishCenter(nodegeom.NodeWorldPos(m.geom))

	if m.tr != nil {
		m.emitGeometry()
	}
}
