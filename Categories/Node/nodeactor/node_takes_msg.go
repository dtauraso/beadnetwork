package nodeactor

import (
	"github.com/dtauraso/wirefold/Categories/Node/movemsg"
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (m *NodeGeometry) take(msg movemsg.Msg) {
	if msg.NodeID != m.id {
		return
	}
	switch b := msg.Body.(type) {
	case movemsg.NeighborMoved:
		m.takeNeighborMove(b)
	case movemsg.Drag:
		m.takeDragOfSelf(b)
	case movemsg.DragStart:
		m.beads.PostBeadDrag(true)
	case movemsg.DragEnd:

		m.beads.PostBeadDrag(false)
	case movemsg.Select:
		m.ui.SetSelected(b.On)
	case movemsg.Hover:
		m.ui.SetHover(b.On, b.Port, b.IsInput)
	case movemsg.Latched:
		m.ui.SetLatched(b.On)
	case movemsg.TiltVectorAngle:
		m.handleTiltVectorAngle(b)
	case movemsg.TiltVectorReset:
		m.handleTiltVectorReset()
	default:
		panic("nodeactor.take: node " + m.id + " was handed a move body it has no case for. " +
			"Every movemsg.Body must be handled here; an unhandled one used to fall through to a " +
			"silent redraw, so the message did nothing and nothing said so.")
	}
}

func (m *NodeGeometry) takeNeighborMove(msg movemsg.NeighborMoved) {
	if msg.Delta != nil && msg.SenderID != "" {
		m.deltas.ShiftOtherBy(msg.SenderID, *msg.Delta)
	}
	if msg.Center != nil {
		idx := polarindex.MeasureIndex(polar.Cart2polarAtTheta(polar.Vec3(msg.Center.Sub(movemsg.Vec3(m.SceneCenter()))), m.ScenePolar().Theta), m.Constants())
		m.ApplyCenter(idx)
		return
	}
	m.emitGeometry()
}

func (m *NodeGeometry) takeDragOfSelf(msg movemsg.Drag) {
	haveIdx := m.ComposedIndex()

	var delta polarindex.Offset
	switch {
	case msg.Target != nil:
		delta = polarindex.Delta(*msg.Target, haveIdx)
	case msg.Delta != nil:
		delta = *msg.Delta
	default:
		panic("nodeactor.takeDragOfSelf: drag of " + m.id + " carries neither a target nor a delta — a pointer drag names WHERE (Target) and the node measures its own delta from its own index; a peer's request names HOW FAR (Delta). One of the two must be set")
	}

	movedIdx := polarindex.Compose(haveIdx, m.TrimOwnDrag(delta), m.Constants())

	m.msg.CommitLocal(m.id, movedIdx)
	newPos := m.SceneCenter().Add(Vec3(polar.Polar2cart(polarindex.ToPolar(movedIdx, m.Constants()))))

	m.writeStreamFrame([]owners.RowEvent{{
		Kind: owners.KindBreadcrumb, Label: BreadcrumbDragCommit, Debug: 1,
		NodeRow: m.stream.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		X: newPos.X, Y: newPos.Y, Z: newPos.Z,
	}})
}

func (m *NodeGeometry) handleTiltVectorAngle(msg movemsg.TiltVectorAngle) {
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
