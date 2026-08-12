// node_geometry_handle.go — NodeGeometry's own inbound-message dispatch: what happens to
// this node's OWN state when a movemsg.Msg of a given Kind arrives. Split out of
// node_geometry.go (docs/planning/movedispatch-decomposition.md §20) — the type, its
// constructor, and this dispatch are two different concerns sharing one composer.
package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// handle applies one move to this node: update held position, re-emit node-geometry.
func (m *NodeGeometry) handle(msg movemsg.Msg) {
	if msg.NodeID != m.id {
		return
	}
	if msg.Kind == movemsg.KindCenter {
		// This node is the SOLE writer of its own position (single-writer by
		// construction — this is the only path that mutates it). A Center payload is
		// the flat absolute-scene-polar drag write from fanCenters: apply it via
		// ApplyCenter, which also re-emits. A nil Center is fanCenters' PARTNER
		// re-emit (a neighbor whose OWN center is unchanged, only asked to re-emit so
		// any reader of its geometry sees a consistent frame) — no mutation, just
		// re-emit.
		if msg.Center != nil {
			m.ApplyCenter(*msg.Center, msg.ReachR)
			return
		}
		if m.tr != nil {
			m.emitGeometry()
		}
		return
	}
	if msg.Kind == movemsg.KindDrag {
		// Owner-goroutine drag entry (generalized to EVERY node so no node's quantized
		// offset is ever touched by a foreign goroutine): commit this node's OWN new
		// position via the local (synchronous-snap-publish) commit path. A drag is
		// always a FREE move now -- there is no equal-radii solve and no propagation
		// past this node's own commit.
		newPos := msg.Target
		if m.msg.commitLocal != nil {
			m.msg.commitLocal(m.id, newPos)
		}
		if m.tr != nil {
			m.tr.Breadcrumb("drag.commit", m.id, "", fmt.Sprintf("newPos=(%.4f,%.4f,%.4f)", newPos.X, newPos.Y, newPos.Z))
			// Structured buffer counterpart, riding this node's own dedicated
			// stream frame (emitGeometry's own next emit already fires from
			// commitLocal above, so this rides as a distinct events-only-shaped
			// write here rather than waiting on that one).
			m.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbDragCommit, Debug: 1,
				NodeRow: m.stream.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				X: newPos.X, Y: newPos.Y, Z: newPos.Z,
			}})
		}
		return
	}
	if msg.Kind == movemsg.KindDragStart {
		m.startBeadDrag()
		return
	}
	if msg.Kind == movemsg.KindDragEnd {
		// "done dragging" is not optional (PLAN.md): sent from gesture.go's
		// gestPointerUp on EVERY path a drag can end by (including one abandoned
		// without a clean pointer-move first — see movemsg.KindDragEnd's own doc
		// comment), so no chain bead this node woke is ever left on machine time.
		m.endBeadDrag()
		return
	}
	if msg.Kind == movemsg.KindSelect {
		if msg.Bool {
			m.ui.selected = 1
		} else {
			m.ui.selected = 0
		}
		return
	}
	if msg.Kind == movemsg.KindHover {
		if msg.Bool {
			m.ui.hovered = 1
			m.ui.hoverPort = msg.Port
			m.ui.hoverIsInput = msg.IsInput
		} else {
			m.ui.hovered = 0
			m.ui.hoverPort = ""
			m.ui.hoverIsInput = false
		}
		return
	}
	if msg.Kind == movemsg.KindLatched {
		if msg.Bool {
			m.ui.latchedSel = 1
		} else {
			m.ui.latchedSel = 0
		}
		return
	}
	if msg.Kind == movemsg.KindTiltVectorAngle {
		// Adjust THIS node's own vector-direction index by one TiltVectorAngleStep click —
		// index arithmetic only (memory/feedback_abc_times_constant_not_rederive.md), no
		// trig here. Persisted immediately to this node's OWN file (persistTiltVectorAngle,
		// quant_offset_persist.go) and re-emitted so the panel's read-only reflect and
		// the drawn arrow both pick up the change on the next frame.
		delta := int32(-1)
		if msg.Bool {
			delta = 1
		}
		m.tilt.topTiltVectorThetaIdx += delta
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: this path only runs for a kind that has NOT claimed BuildArgs.TiltEditIn
		// (every kind except PairNode today — see movemsg.KindTiltVectorAngle's own doc
		// comment and applyUpdateTiltVector's fallback, package Wiring's stdin_reader.go).
		// PairNode's own tilt-panel edits are routed to its OWN goroutine instead
		// (TiltEditIn), which applies the click, syncs this value back via
		// PairNodeSelf.SetTiltIndex, AND places "the kick" bead on its own Out directly
		// — none of that happens here anymore.
		return
	}
	if msg.Kind == movemsg.KindTiltVectorReset {
		// Return THIS node's own vector direction to the start position — both indices to
		// 0, the documented default (tilt vector at world +y). No bead: this is a
		// stop-and-return, not a kick. Persisted immediately, same as an adjust.
		m.tilt.topTiltVectorThetaIdx = 0
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: same split as movemsg.KindTiltVectorAngle — this path only runs for a kind
		// that has NOT claimed BuildArgs.TiltEditIn. PairNode routes a reset through
		// its own TiltEditIn/movemsg.TiltEditMsg.Reset instead.
		return
	}
	// movemsg.KindTiltIndexSync/ReceivedVectorSync/BeadClear are GONE
	// (task/pair-node-owns-itself): a pair node (PairNode) owns this geometry
	// directly (PairNodeSelf, pair_node_self.go), so what used to be a one-way
	// notification message to itself is now a plain method call on the same
	// goroutine — see PairNodeSelf.SetTiltIndex/SetReceivedVector/ClearOutBeads,
	// which apply exactly what this handle() branch used to apply.
	if msg.Kind == movemsg.KindNeighborCenter {
		// Delivery-mechanism push (see ApplyCenter/partnerCenters' doc comments): a
		// direct neighbor's OWN center just changed. Store it in THIS node's owned
		// partnerCenters map (write, own goroutine only) and re-emit THIS node's own
		// geometry so its aimed ports pick up the fresh partner center — same value,
		// same effect as the old cross-goroutine snap read, just message-delivered.
		// ONE HOP ONLY: this node's own center did NOT change, so it must never push
		// a NeighborCenter of its own onward from here (no cascade past this point).
		if m.topo.partnerCenters == nil {
			m.topo.partnerCenters = map[string]vec3{}
		}
		m.topo.partnerCenters[msg.SenderID] = msg.FromCenter
		if m.tr != nil {
			// DIAGNOSTIC ONLY (task/log-node4-chain-aim): records that this node's own
			// goroutine received a neighbor-center push, and from whom, so a drag-time
			// trace can show whether/when it arrives relative to this node's own emits.
			value := fmt.Sprintf("sender=%s center=(%.4f,%.4f,%.4f)", msg.SenderID, msg.FromCenter.X, msg.FromCenter.Y, msg.FromCenter.Z)
			m.tr.Breadcrumb("neighbor-center-recv", m.id, msg.SenderID, value)
			senderRow := int32(-1)
			if m.topo.nodeRowFor != nil {
				if r, ok := m.topo.nodeRowFor(msg.SenderID); ok {
					senderRow = r
				}
			}
			m.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbNeighborCenterRecv, Debug: 1,
				NodeRow: m.stream.nodeRow, PortRow: -1, TargetRow: senderRow, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Text: value,
			}})
			m.emitGeometry()
		}
		return
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}
