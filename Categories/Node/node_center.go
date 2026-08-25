package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (m *NodeGeometry) ApplyCenter(idx polarindex.Index) {
	SetNodeWorld(&m.geom, idx)

	m.msg.PublishCenter(Vec3(NodeWorldPos(m.geom)))

	m.emitGeometry()
}
