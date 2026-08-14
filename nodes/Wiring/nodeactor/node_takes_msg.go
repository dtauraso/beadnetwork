package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
)

// take is one message off THIS node's own inbox, on its own goroutine. It is
// not a router standing between senders and nodes: it is unexported, its only
// call site is this node's mover loop (NodeMover.Run), and every arm below is a
// method on `m`. A message addressed to someone else is dropped rather than
// forwarded, because there is nobody here to forward it on behalf of.
//
// The arms are deliberately not named alike. The first two are this node
// ANSWERING about its own position — it decides what it takes of a drag of
// itself, and it takes a neighbour's move onto its own side of their edge. The
// rest set a flag from the message and have no answer to give, so they stay
// spelled as handling.
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

// takeNeighborMove is what this node does when the node at the other end of an
// edge has moved and said how far. It is told a Δ, never a position, and it
// applies that Δ to its own side of that edge — see movemsg.Msg.Delta.
func (m *NodeGeometry) takeNeighborMove(msg movemsg.Msg) {
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

// takeDragOfSelf is this node's answer to being dragged. A pointer ASKS, as a
// polar Δ against this node's own centre; how much of it the node takes is its
// own (TrimOwnDrag), decided here, on its own goroutine, out of its own state.
func (m *NodeGeometry) takeDragOfSelf(msg movemsg.Msg) {
	newPos := msg.Target
	targetPolar := msg.TargetPolar

	// A drag of THIS node arrives as the polar delta triple it was converted
	// to, and this node decides how much of it to take. The trim is its own
	// (TrimOwnDrag) and runs HERE, on its own goroutine, holding its own rules
	// — a drag is a request, and the answer to it is the node's.
	//
	// Composing is the node's too, for the same reason the trim is: the three
	// numbers are added to its own point and the result is what commits. Only
	// then does the triple become a world position, once, for whoever needs to
	// draw it.
	if msg.Delta != nil {
		moved := polar.Compose(m.ScenePolar(), m.TrimOwnDrag(*msg.Delta))
		targetPolar = &moved
		newPos = m.SceneCenter().Add(polar.Polar2cart(moved))
	}

	// The triple, when the sender composed one, is what commits — see
	// movemsg.Msg.TargetPolar for what the world does to a triple it carries.
	// The load-time hold is the sender that still composes: it states a point
	// absolutely, which is the one thing a delta cannot do.
	m.msg.CommitLocal(m.id, newPos, targetPolar)
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
