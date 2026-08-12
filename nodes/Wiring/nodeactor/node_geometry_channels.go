package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

func (m *NodeGeometry) NeighborTrySend(fromID string) (func(movemsg.Msg) bool, bool) {
	return m.msg.NeighborTrySend(fromID)
}

func (m *NodeGeometry) PollCenter() (vec3, bool) {
	return m.msg.PollCenter()
}

func (m *NodeGeometry) SendExternal(ctx context.Context, msg movemsg.Msg) {
	m.msg.SendExternal(ctx, msg)
}

func (m *NodeGeometry) TryRecvExternal() (movemsg.Msg, bool) {
	return m.msg.TryRecvExternal()
}

func (m *NodeGeometry) EnqueueSend(destID string, msg movemsg.Msg) {
	m.msg.EnqueueSend(m.id, destID, msg)
}
