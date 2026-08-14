package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
)

func (m *NodeGeometry) handle(msg movemsg.Msg) {
	if msg.NodeID != m.id {
		return
	}
	switch msg.Kind {
	case movemsg.KindCenter:
		m.handleCenter(msg)
	case movemsg.KindDrag:
		m.handleDrag(msg)
	case movemsg.KindDragStart:
		m.beads.StartBeadDrag()
	case movemsg.KindDragEnd:

		m.beads.EndBeadDrag()
	case movemsg.KindSelect:
		m.handleSelect(msg)
	case movemsg.KindHover:
		m.handleHover(msg)
	case movemsg.KindLatched:
		m.handleLatched(msg)
	case movemsg.KindTiltVectorAngle:
		m.handleTiltVectorAngle(msg)
	case movemsg.KindTiltVectorReset:
		m.handleTiltVectorReset()
	default:

		if m.tr != nil {
			m.emitGeometry()
		}
	}
}

func (m *NodeGeometry) handleCenter(msg movemsg.Msg) {
	// The node at the other end of an edge moved and said how far. This node
	// did not move, so its side of that edge gains the whole of that Δ.
	if msg.Delta != nil && msg.SenderID != "" {
		m.deltas.ShiftOtherBy(msg.SenderID, *msg.Delta)
	}
	if msg.Center != nil {
		m.ApplyCenter(*msg.Center, msg.ReachR)
		return
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

func (m *NodeGeometry) handleDrag(msg movemsg.Msg) {
	newPos := msg.Target
	// The triple, when the sender composed one, is what commits — see
	// movemsg.Msg.TargetPolar for what the world does to a triple it carries.
	m.msg.CommitLocal(m.id, newPos, msg.TargetPolar)
	if m.tr != nil {
		m.tr.Breadcrumb("drag.commit", m.id, "", fmt.Sprintf("newPos=(%.4f,%.4f,%.4f)", newPos.X, newPos.Y, newPos.Z))

		m.writeStreamFrame([]rowevent.RowEvent{{
			Kind: T.KindBreadcrumb, Label: T.BreadcrumbDragCommit, Debug: 1,
			NodeRow: m.stream.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			X: newPos.X, Y: newPos.Y, Z: newPos.Z,
		}})
	}
}

func (m *NodeGeometry) handleSelect(msg movemsg.Msg) {
	m.ui.SetSelected(msg.Bool)
}

func (m *NodeGeometry) handleHover(msg movemsg.Msg) {
	m.ui.SetHover(msg.Bool, msg.Port, msg.IsInput)
}

func (m *NodeGeometry) handleLatched(msg movemsg.Msg) {
	m.ui.SetLatched(msg.Bool)
}

func (m *NodeGeometry) handleTiltVectorAngle(msg movemsg.Msg) {
	delta := int32(-1)
	if msg.Bool {
		delta = 1
	}
	m.tilt.BumpTopTiltVectorThetaIdx(delta)
	m.persistTiltVectorAngle()
	if m.tr != nil {
		m.emitGeometry()
	}
}

func (m *NodeGeometry) handleTiltVectorReset() {
	m.tilt.ResetTopTiltVectorThetaIdx()
	m.persistTiltVectorAngle()
	if m.tr != nil {
		m.emitGeometry()
	}
}
