package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

func (m *NodeGeometry) ApplyCenter(center vec3, reach float64) {
	// This node moved; its neighbours did not. Rebase every stored path by
	// the node's own delta so each still lands on the same neighbour.
	prev := nodegeom.NodeWorldPos(m.geom)
	nodegeom.SetNodeWorld(&m.geom, center)
	m.topo.RebaseForSelfMove(prev, center)
	m.geom.ReachR = reach

	m.msg.BroadcastCenter(m.id, center)

	if m.tr != nil {
		m.emitGeometry()
	}
}
