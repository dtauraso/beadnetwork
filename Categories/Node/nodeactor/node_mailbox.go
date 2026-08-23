package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
)

func (m *NodeGeometry) NeighborDeposit(fromID string) (owners.Deposit, bool) {
	return m.msg.NeighborDeposit(fromID)
}

func (m *NodeGeometry) PollCenter() (Vec3, bool) {
	c, ok := m.msg.PollCenter()
	return Vec3(c), ok
}

func (m *NodeGeometry) SendExternal(ctx context.Context, msg owners.Msg) {
	m.msg.SendExternal(ctx, msg)
}

func (m *NodeGeometry) EnqueueSend(destID string, msg owners.Msg) {
	m.msg.EnqueueSend(m.id, destID, msg)
}
