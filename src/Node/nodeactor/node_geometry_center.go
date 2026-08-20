package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
)

func (m *NodeGeometry) ApplyCenter(idx polarindex.Index) {
	nodegeom.SetNodeWorld(&m.geom, idx)

	m.msg.PublishCenter(nodegeom.NodeWorldPos(m.geom))

	m.emitGeometry()
}
