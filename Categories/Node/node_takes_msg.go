package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/Geom"
	"github.com/dtauraso/beadnetwork/Categories/Node/Messaging"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

const BreadcrumbDragCommit = "drag-commit"

func (m *NodeGeometry) take(msg Messaging.Msg) {
	if msg.NodeID != m.id {
		return
	}
	switch b := msg.Body.(type) {
	case Messaging.NeighborMoved:
		m.takeNeighborMove(b)
	case Messaging.Drag:
		m.takeDragOfSelf(b)
	case Messaging.DragStart:
		m.beads.PostBeadDrag(true)
	case Messaging.DragEnd:

		m.beads.PostBeadDrag(false)
	case Messaging.Select:
		m.ui.SetSelected(b.On)
	case Messaging.Hover:
		m.ui.SetHover(b.On, b.Port, b.IsInput)
	case Messaging.Latched:
		m.ui.SetLatched(b.On)
	default:
		panic("Node.take: node " + m.id + " was handed a move body it has no case for. " +
			"Every Body must be handled here; an unhandled one used to fall through to a " +
			"silent redraw, so the message did nothing and nothing said so.")
	}
}

func (m *NodeGeometry) takeNeighborMove(msg Messaging.NeighborMoved) {
	if msg.Delta != nil && msg.SenderID != "" {
		m.deltas.ShiftOtherBy(msg.SenderID, *msg.Delta)
	}
	if msg.Center != nil {
		idx := Geom.IndexAtTheta(m.geom.SceneCenter, Vec3(*msg.Center), m.ScenePolar().Theta, m.Constants())
		m.ApplyCenter(idx)
		return
	}
	m.emitGeometry()
}

func (m *NodeGeometry) takeDragOfSelf(msg Messaging.Drag) {
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
	newPos := Geom.WorldPosOf(m.geom, movedIdx)

	m.writeStreamFrame([]RowEvent{{
		Kind: KindBreadcrumb, Label: BreadcrumbDragCommit, Debug: 1,
		NodeRow: m.stream.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		X: newPos.X, Y: newPos.Y, Z: newPos.Z,
	}})
}
