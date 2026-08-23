package Node

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Node/owners"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (m *NodeGeometry) ApplyCenter(idx polarindex.Index) {
	nodegeom.SetNodeWorld(&m.geom, idx)

	m.msg.PublishCenter(owners.Vec3(nodegeom.NodeWorldPos(m.geom)))

	m.emitGeometry()
}
