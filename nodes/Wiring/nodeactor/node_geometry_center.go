package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

func (m *NodeGeometry) ApplyCenter(center vec3, reach float64) {
	nodegeom.SetNodeWorld(&m.geom, center)
	m.geom.ReachR = reach

	m.msg.PublishCenter(center)

	if m.tr != nil {
		m.emitGeometry()
	}
}
