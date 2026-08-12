package nodeactor

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

func (m *NodeGeometry) NeighborTrySend(fromID string) (func(movemsg.Msg) bool, bool) {
	ch, ok := m.msg.neighborIn[fromID]
	if !ok {
		return nil, false
	}
	return func(msg movemsg.Msg) bool {
		select {
		case ch <- msg:
			return true
		default:
			return false
		}
	}, true
}

func (m *NodeGeometry) PollCenter() (vec3, bool) {
	select {
	case c := <-m.msg.centerOut:
		return c, true
	default:
		return vec3{}, false
	}
}

func (m *NodeGeometry) SendExternal(ctx context.Context, msg movemsg.Msg) {
	if ctx == nil {
		m.msg.extIn <- msg
		return
	}
	select {
	case m.msg.extIn <- msg:
	case <-ctx.Done():
	}
}

func (m *NodeGeometry) TryRecvExternal() (movemsg.Msg, bool) {
	select {
	case msg := <-m.msg.extIn:
		return msg, true
	default:
		return movemsg.Msg{}, false
	}
}

func (m *NodeGeometry) EnqueueSend(destID string, msg movemsg.Msg) {
	m.msg.pending = append(m.msg.pending, pendingSend{destID: destID, msg: msg})
	m.flushPending()
	if len(m.msg.pending) > maxPendingSends {

		panic(fmt.Sprintf(
			"NodeGeometry(%s): pending exceeded %d retry-queued sends; either a "+
				"destination's own goroutine has stopped draining its inbox "+
				"(wedged or dead), or this node is enqueueing to a peer faster "+
				"than that peer drains, cycle over cycle",
			m.id, maxPendingSends))
	}
}
