package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"

	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

func (m *NodeGeometry) take(msg movemsg.Msg) {
	if msg.NodeID != m.id {
		return
	}
	switch msg.Kind {
	case movemsg.KindCenter:
		m.takeNeighborMove(msg)
	case movemsg.KindDrag:
		m.takeDragOfSelf(msg)
	case movemsg.KindDragStart:
		m.beads.PostBeadDrag(true)
	case movemsg.KindDragEnd:

		m.beads.PostBeadDrag(false)
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

		m.emitGeometry()
	}
}

func (m *NodeGeometry) takeNeighborMove(msg movemsg.Msg) {
	if msg.Delta != nil && msg.SenderID != "" {
		m.deltas.ShiftOtherBy(msg.SenderID, *msg.Delta)
	}
	if msg.Center != nil {
		idx := polarindex.MeasureIndex(polar.Cart2polarAtTheta(msg.Center.Sub(m.SceneCenter()), m.ScenePolar().Theta), m.Constants())
		m.ApplyCenter(idx)
		return
	}
	m.emitGeometry()
}

func (m *NodeGeometry) takeDragOfSelf(msg movemsg.Msg) {
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
	newPos := m.SceneCenter().Add(polar.Polar2cart(polarindex.ToPolar(movedIdx, m.Constants())))

	m.writeStreamFrame([]B.RowEvent{{
		Kind: B.KindBreadcrumb, Label: B.BreadcrumbDragCommit, Debug: 1,
		NodeRow: m.stream.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		X: newPos.X, Y: newPos.Y, Z: newPos.Z,
	}})
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
	m.tilt.BumpTopTiltVectorPhiIdx(delta)
	m.persistTiltVectorAngle()
	m.emitGeometry()
}

func (m *NodeGeometry) handleTiltVectorReset() {
	m.tilt.ResetTopTiltVectorPhiIdx()
	m.persistTiltVectorAngle()
	m.emitGeometry()
}
