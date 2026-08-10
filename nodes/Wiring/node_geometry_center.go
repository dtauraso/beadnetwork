// node_geometry_center.go — the sole write of a nodeGeometry's own center/reach.
// PerpendicularThetaIdx (the constant this write's tilt-vector-angle callers compare
// against) moved to nodes/Wiring/tiltvector alongside the rest of the tilt-vector-channel
// vocabulary (god-object decomposition, pure move — no logic change).
package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

// applyCenter is the SOLE WRITE of this node's center/reach. It is called ONLY from this
// node's own driving goroutine (handle's movemsg.KindCenter case, driven by fanCenters
// below), which is what makes that one goroutine the exclusive writer of m.geom. It sets
// the held polar position, pushes the fresh center to the dispatch goroutine's owned
// center mirror (m.msg.centerOut, latest-wins — see its doc comment) and to every direct
// neighbor's partnerCenters map (below), and re-emits this node's live geometry.
func (m *nodeGeometry) applyCenter(center vec3, reach float64) {
	nodegeom.SetNodeWorld(&m.geom, center)
	m.geom.ReachR = reach
	// Latest-wins non-blocking push onto centerOut: drain any stale unread value first
	// so the slot always ends up holding the newest center, never blocking this
	// goroutine even if the dispatch goroutine hasn't drained the previous push yet.
	select {
	case <-m.msg.centerOut:
	default:
	}
	select {
	case m.msg.centerOut <- center:
	default:
	}
	// Push this fresh center to every direct neighbor (nm.msg.neighborIn's key set — one
	// hop, no cascade) so each neighbor's OWN partnerCenters map picks it up via
	// movemsg.KindNeighborCenter (handle, below). Routed through m.msg.sendMove (this
	// node's own retry queue), same as every other fan-out this file makes, so a
	// momentarily-full neighbor inbox is retried, never dropped or blocking. Sent
	// BEFORE this same commit's broadcastToEdgesAndPartners nil-Center re-emit (called
	// right after applyCenter by every live caller), so per-destination FIFO delivers
	// this push first and the re-emit always sees the just-pushed center.
	for neighborID := range m.msg.neighborIn {
		m.msg.sendMove(neighborID, movemsg.Msg{Kind: movemsg.KindNeighborCenter, NodeID: neighborID,
			SenderID: m.id, FromCenter: center})
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}
