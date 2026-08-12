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
		m.beads.startBeadDrag()
	case movemsg.KindDragEnd:

		m.beads.endBeadDrag()
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
	case movemsg.KindNeighborCenter:
		m.handleNeighborCenter(msg)
	default:

		if m.tr != nil {
			m.emitGeometry()
		}
	}
}

func (m *NodeGeometry) handleCenter(msg movemsg.Msg) {
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
	if m.msg.commitLocal != nil {
		m.msg.commitLocal(m.id, newPos)
	}
	if m.tr != nil {
		m.tr.Breadcrumb("drag.commit", m.id, "", fmt.Sprintf("newPos=(%.4f,%.4f,%.4f)", newPos.X, newPos.Y, newPos.Z))

		m.writeStreamFrame([]rowevent.RowEvent{{
			Kind: T.KindBreadcrumb, Label: T.BreadcrumbDragCommit, Debug: 1,
			NodeRow: m.stream.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
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

func (m *NodeGeometry) handleNeighborCenter(msg movemsg.Msg) {
	m.topo.SeedPartnerCenter(msg.SenderID, msg.FromCenter)
	if m.tr != nil {

		value := fmt.Sprintf("sender=%s center=(%.4f,%.4f,%.4f)", msg.SenderID, msg.FromCenter.X, msg.FromCenter.Y, msg.FromCenter.Z)
		m.tr.Breadcrumb("neighbor-center-recv", m.id, msg.SenderID, value)
		senderRow := int32(-1)
		if m.topo.nodeRowFor != nil {
			if r, ok := m.topo.nodeRowFor(msg.SenderID); ok {
				senderRow = r
			}
		}
		m.writeStreamFrame([]rowevent.RowEvent{{
			Kind: T.KindBreadcrumb, Label: T.BreadcrumbNeighborCenterRecv, Debug: 1,
			NodeRow: m.stream.nodeRow, PortRow: -1, TargetRow: senderRow, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Text: value,
		}})
		m.emitGeometry()
	}
}
