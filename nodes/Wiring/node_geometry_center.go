// node_geometry_center.go — the sole write of a nodeGeometry's own center/reach, and the
// one constant that write's tilt-vector-angle callers compare against.
package Wiring

// PerpendicularThetaIdx is the topTiltVectorThetaIdx value at which the tilt vector is exactly
// perpendicular to world +y: CurveParamTiltVectorAngleStep is π/12 (15°), and π/2 (90°) is
// exactly 6 steps. Comparing to this INTEGER is what makes the straightening loop's stop
// condition exact — cos(π/2) in float64 is ~6.1e-17, so a literal float dot==0 test would
// never fire (memory/feedback_abc_times_constant_not_rederive.md: index arithmetic, trig
// only at the cartesian/polar boundary). Exported (capitalized) so PairNode's own
// goroutine — which now runs the straightening rule itself, per-package — can compare
// against it without duplicating the constant; the rule itself no longer lives here (see
// nodes/PairNode/node.go).
//
// dot(tilt, coplanarNormal) == 0 is decided as thetaIdx == PerpendicularThetaIdx, not by
// computing an actual float dot product. STATE THE ASSUMPTION THAT MAKES THE SHORTCUT
// VALID: the tilt vector's in-plane angle IS its θ index only because, for this scene, the
// ring plane holds world +y and θ is measured from +y, so the two coincide (see
// topTiltVectorThetaIdx's own doc comment and the CoplanarNormal/UpAxis derivations in
// writeStreamFrame above). A scene whose ring plane does NOT contain +y breaks that
// coincidence — θ would then measure something unrelated to the coplanar normal, and the
// rule would need to compare an actual dot(tilt, coplanarNormal) via the two integer
// indices' angles converted through anglesToWorldOffset, not thetaIdx alone.
const PerpendicularThetaIdx int32 = 6

// applyCenter is the SOLE WRITE of this node's center/reach. It is called ONLY from this
// node's own driving goroutine (handle's moveMsgKindCenter case, driven by fanCenters
// below), which is what makes that one goroutine the exclusive writer of m.geom. It sets
// the held polar position, pushes the fresh center to the dispatch goroutine's owned
// center mirror (m.msg.centerOut, latest-wins — see its doc comment) and to every direct
// neighbor's partnerCenters map (below), and re-emits this node's live geometry.
func (m *nodeGeometry) applyCenter(center vec3, reach float64) {
	setNodeWorld(&m.geom, center)
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
	// moveMsgKindNeighborCenter (handle, below). Routed through m.msg.sendMove (this
	// node's own retry queue), same as every other fan-out this file makes, so a
	// momentarily-full neighbor inbox is retried, never dropped or blocking. Sent
	// BEFORE this same commit's broadcastToEdgesAndPartners nil-Center re-emit (called
	// right after applyCenter by every live caller), so per-destination FIFO delivers
	// this push first and the re-emit always sees the just-pushed center.
	for neighborID := range m.msg.neighborIn {
		m.msg.sendMove(neighborID, moveMsg{Kind: moveMsgKindNeighborCenter, NodeID: neighborID,
			SenderID: m.id, FromCenter: center})
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}
