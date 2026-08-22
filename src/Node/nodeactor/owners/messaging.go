package owners

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
)

type Messaging struct {
	extIn chan movemsg.Msg

	dragIn *neighborSlot

	neighborIn map[string]*neighborSlot

	centerOut chan Vec3

	sendMove func(id string, msg movemsg.Msg)

	resolveDest func(id string) (Deposit, bool)

	commitLocal func(id string, idx polarindex.Index)
}

type Deposit func(msg movemsg.Msg)

func NewMessaging(extIn chan movemsg.Msg, centerOut chan Vec3) Messaging {
	return Messaging{
		extIn:      extIn,
		dragIn:     newNeighborSlot(),
		neighborIn: map[string]*neighborSlot{},
		centerOut:  centerOut,
	}
}

func (n *Messaging) WireMessaging(
	resolveDest func(id string) (Deposit, bool),
	sendMove func(id string, msg movemsg.Msg),
	commitLocal func(id string, idx polarindex.Index),
) {
	n.resolveDest = resolveDest
	n.sendMove = sendMove
	n.commitLocal = commitLocal
}

func (n *Messaging) EnsureNeighborChannel(otherID string) {
	if _, exists := n.neighborIn[otherID]; !exists {
		n.neighborIn[otherID] = newNeighborSlot()
	}
}

func (n *Messaging) SendMove() func(id string, msg movemsg.Msg) { return n.sendMove }

func (n *Messaging) SeedCenter(center Vec3) {
	n.centerOut <- center
}

func (n *Messaging) CommitLocal(id string, idx polarindex.Index) {
	if n.commitLocal != nil {
		n.commitLocal(id, idx)
	}
}

func (n *Messaging) DrainPending(ctx context.Context, handle func(movemsg.Msg)) (progressed, cancelled bool) {
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
	if msg, ok := n.dragIn.take(); ok {
		handle(msg)
		progressed = true
	}
	for _, slot := range n.neighborIn {
		if msg, ok := slot.take(); ok {
			handle(msg)
			if msg.TestDone != nil {
				close(msg.TestDone)
			}
			progressed = true
		}
	}
	return progressed, false
}

func (n *Messaging) NeighborIDs() []string {
	ids := make([]string, 0, len(n.neighborIn))
	for id := range n.neighborIn {
		ids = append(ids, id)
	}
	return ids
}

func (n *Messaging) NeighborDeposit(fromID string) (Deposit, bool) {
	slot, ok := n.neighborIn[fromID]
	if !ok {
		return nil, false
	}
	return slot.deposit, true
}

func (n *Messaging) PollCenter() (Vec3, bool) {
	select {
	case c := <-n.centerOut:
		return c, true
	default:
		return Vec3{}, false
	}
}

func (n *Messaging) SendExternal(_ context.Context, msg movemsg.Msg) {
	if msg.Kind == movemsg.KindDrag {
		n.dragIn.deposit(msg)
		return
	}
	select {
	case n.extIn <- msg:
	default:
		panic(fmt.Sprintf(
			"NodeGeometry(%s): discrete-event inbox full at %d unread (kind %q); these are "+
				"human decision-rate events (select/hover/dragStart/dragEnd/tilt) and cannot "+
				"outrun a geometry loop that runs every real tick — so either this node's "+
				"geometry goroutine has stopped running, or a continuous per-pointer-move "+
				"quantity is being sent here instead of onto a coalescing slot",
			msg.NodeID, inboxDepth, msg.Kind))
	}
}

func (n *Messaging) TryRecvExternal() (movemsg.Msg, bool) {
	select {
	case msg := <-n.extIn:
		return msg, true
	default:
		return movemsg.Msg{}, false
	}
}

func (n *Messaging) EnqueueSend(_, destID string, msg movemsg.Msg) {
	if n.resolveDest == nil {
		return
	}
	deposit, ok := n.resolveDest(destID)
	if !ok {
		return
	}
	deposit(msg)
}

func (n *Messaging) PublishCenter(center Vec3) {
	select {
	case <-n.centerOut:
	default:
	}
	select {
	case n.centerOut <- center:
	default:
	}
}
