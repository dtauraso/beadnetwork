package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

func (m *NodeGeometry) ApplyCenter(center vec3, reach float64) {
	nodegeom.SetNodeWorld(&m.geom, center)
	m.geom.ReachR = reach

	select {
	case <-m.msg.centerOut:
	default:
	}
	select {
	case m.msg.centerOut <- center:
	default:
	}

	for neighborID := range m.msg.neighborIn {
		m.msg.sendMove(neighborID, movemsg.Msg{Kind: movemsg.KindNeighborCenter, NodeID: neighborID,
			SenderID: m.id, FromCenter: center})
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}
