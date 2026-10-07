package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

const BreadcrumbDragCommit = "drag-commit"

func (m *NodeGeometry) take(msg Msg) {
	if msg.NodeID != m.id {
		return
	}
	switch b := msg.Body.(type) {
	case NeighborMoved:
		m.takeNeighborMove(b)
	case Drag:
		m.takeDragOfSelf(b)
	case DragStart:
		m.beads.PostBeadDrag(true)
	case DragEnd:

		m.beads.PostBeadDrag(false)
	case Select:
		m.ui.SetSelected(b.On)
	case Hover:
		m.ui.SetHover(b.On, b.Port, b.IsInput)
	case Latched:
		m.ui.SetLatched(b.On)
	case TiltVectorAngle:
		m.handleTiltVectorAngle(b)
	case TiltVectorReset:
		m.handleTiltVectorReset()
	default:
		panic("Node.take: node " + m.id + " was handed a move body it has no case for. " +
			"Every Body must be handled here; an unhandled one used to fall through to a " +
			"silent redraw, so the message did nothing and nothing said so.")
	}
}

func (m *NodeGeometry) takeNeighborMove(msg NeighborMoved) {
	if msg.Delta != nil && msg.SenderID != "" {
		m.deltas.ShiftOtherBy(msg.SenderID, *msg.Delta)
	}
	if msg.Center != nil {
		idx := IndexAtTheta(m.geom.SceneCenter, Vec3(*msg.Center), m.ScenePolar().Theta, m.Constants())
		m.ApplyCenter(idx)
		return
	}
	m.emitGeometry()
}

func (m *NodeGeometry) takeDragOfSelf(msg Drag) {
	haveIdx := m.ComposedIndex()

	var delta polarindex.Offset
	switch {
	case msg.Target != nil:
		delta = polarindex.Delta(*msg.Target, haveIdx)
	case msg.Delta != nil:
		delta = *msg.Delta
	default:
		panic("Node.takeDragOfSelf: drag of " + m.id + " carries neither a target nor a delta — a pointer drag names WHERE (Target) and the node measures its own delta from its own index; a peer's request names HOW FAR (Delta). One of the two must be set")
	}

	movedIdx := polarindex.Compose(haveIdx, m.TrimOwnDrag(delta), m.Constants())

	m.msg.CommitLocal(m.id, movedIdx)
	newPos := WorldPosAt(m.geom.SceneCenter, movedIdx, m.Constants())

	m.writeStreamFrame([]RowEvent{{
		Kind: KindBreadcrumb, Label: BreadcrumbDragCommit, Debug: 1,
		NodeRow: m.stream.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		X: newPos.X, Y: newPos.Y, Z: newPos.Z,
	}})
}

func (m *NodeGeometry) handleTiltVectorAngle(msg TiltVectorAngle) {
	delta := int32(-1)
	if msg.Up {
		delta = 1
	}
	m.tilt.BumpTopTiltVectorPhiIdx(delta)
	m.persistTiltVectorAngle()
	m.emitGeometry()
}

func (m *NodeGeometry) handleTiltVectorReset() {
	m.tilt.ResetTopTiltVectorPhiIdx()
	m.persistTiltVectorAngle()
	m.emitGeometry()
}
