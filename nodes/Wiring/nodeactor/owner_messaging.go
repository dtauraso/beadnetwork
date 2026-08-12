package nodeactor

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

func (n *nodeMessaging) WireMessaging(
	resolveDest func(id string) (func(movemsg.Msg) bool, bool),
	sendMove func(id string, msg movemsg.Msg),
	centerOf func(id string) (vec3, bool),
	commitLocal func(id string, newPos vec3),
) {
	n.resolveDest = resolveDest
	n.sendMove = sendMove
	n.centerOf = centerOf
	n.commitLocal = commitLocal
}

func (n *nodeMessaging) EnsureNeighborChannel(otherID string) {
	if _, exists := n.neighborIn[otherID]; !exists {
		n.neighborIn[otherID] = make(chan movemsg.Msg, inboxDepth)
	}
}

func (n *nodeMessaging) SendMove() func(id string, msg movemsg.Msg) { return n.sendMove }

func (n *nodeMessaging) SeedCenter(center vec3) {
	n.centerOut <- center
}

func (n *nodeMessaging) CommitLocal(id string, newPos vec3) {
	if n.commitLocal != nil {
		n.commitLocal(id, newPos)
	}
}

func (n *nodeMessaging) DrainPending(ctx context.Context, handle func(movemsg.Msg)) (progressed, cancelled bool) {
	select {
	case <-ctx.Done():
		return false, true
	case msg := <-n.extIn:
		handle(msg)
		if msg.TestDone != nil {
			close(msg.TestDone)
		}
		progressed = true
	default:
	}
	for _, ch := range n.neighborIn {
		select {
		case msg := <-ch:
			handle(msg)
			if msg.TestDone != nil {
				close(msg.TestDone)
			}
			progressed = true
		default:
		}
	}
	return progressed, false
}

func (n *nodeMessaging) NeighborIDs() []string {
	ids := make([]string, 0, len(n.neighborIn))
	for id := range n.neighborIn {
		ids = append(ids, id)
	}
	return ids
}

func (n *nodeMessaging) NeighborTrySend(fromID string) (func(movemsg.Msg) bool, bool) {
	ch, ok := n.neighborIn[fromID]
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

func (n *nodeMessaging) PollCenter() (vec3, bool) {
	select {
	case c := <-n.centerOut:
		return c, true
	default:
		return vec3{}, false
	}
}

func (n *nodeMessaging) SendExternal(ctx context.Context, msg movemsg.Msg) {
	if ctx == nil {
		n.extIn <- msg
		return
	}
	select {
	case n.extIn <- msg:
	case <-ctx.Done():
	}
}

func (n *nodeMessaging) TryRecvExternal() (movemsg.Msg, bool) {
	select {
	case msg := <-n.extIn:
		return msg, true
	default:
		return movemsg.Msg{}, false
	}
}

func (n *nodeMessaging) EnqueueSend(id, destID string, msg movemsg.Msg) {
	n.pending = append(n.pending, pendingSend{destID: destID, msg: msg})
	n.FlushPending()
	if len(n.pending) > maxPendingSends {

		panic(fmt.Sprintf(
			"NodeGeometry(%s): pending exceeded %d retry-queued sends; either a "+
				"destination's own goroutine has stopped draining its inbox "+
				"(wedged or dead), or this node is enqueueing to a peer faster "+
				"than that peer drains, cycle over cycle",
			id, maxPendingSends))
	}
}

func (n *nodeMessaging) FlushPending() {
	if len(n.pending) == 0 || n.resolveDest == nil {
		return
	}
	blocked := map[string]bool{}
	kept := n.pending[:0]
	for _, item := range n.pending {
		if blocked[item.destID] {
			kept = append(kept, item)
			continue
		}
		trySend, ok := n.resolveDest(item.destID)
		if !ok {
			continue
		}
		if !trySend(item.msg) {
			blocked[item.destID] = true
			kept = append(kept, item)
		}
	}
	n.pending = kept
}

func (n *nodeMessaging) BroadcastCenter(selfID string, center vec3) {
	select {
	case <-n.centerOut:
	default:
	}
	select {
	case n.centerOut <- center:
	default:
	}

	for neighborID := range n.neighborIn {
		n.sendMove(neighborID, movemsg.Msg{Kind: movemsg.KindNeighborCenter, NodeID: neighborID,
			SenderID: selfID, FromCenter: center})
	}
}
