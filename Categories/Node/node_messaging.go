package Node

import (
	"context"
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Messaging struct {
	extIn chan Msg

	dragIn *neighborSlot

	neighborIn map[string]*neighborSlot

	centerOut chan Vec3

	sendMove func(id string, msg Msg)

	resolveDest func(id string) (Deposit, bool)

	commitLocal func(id string, idx polarindex.Index)
}

type Deposit func(msg Msg)

func NewMessaging(extIn chan Msg, centerOut chan Vec3) Messaging {
	return Messaging{
		extIn:      extIn,
		dragIn:     newNeighborSlot(),
		neighborIn: map[string]*neighborSlot{},
		centerOut:  centerOut,
	}
}

func (n *Messaging) WireMessaging(
	resolveDest func(id string) (Deposit, bool),
	sendMove func(id string, msg Msg),
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

func (n *Messaging) SendMove() func(id string, msg Msg) { return n.sendMove }

func (n *Messaging) SeedCenter(center Vec3) {
	n.centerOut <- center
}

func (n *Messaging) CommitLocal(id string, idx polarindex.Index) {
	if n.commitLocal != nil {
		n.commitLocal(id, idx)
	}
}

func (n *Messaging) DrainPending(ctx context.Context, handle func(Msg)) (progressed, cancelled bool) {
	select {
	case <-ctx.Done():
		return false, true
	case msg := <-n.extIn:
		handle(msg)
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

func (n *Messaging) SendExternal(_ context.Context, msg Msg) {
	if _, isDrag := msg.Body.(Drag); isDrag {
		n.dragIn.deposit(msg)
		return
	}
	select {
	case n.extIn <- msg:
	default:
		panic(fmt.Sprintf(
			"NodeGeometry(%s): discrete-event inbox full at %d unread (body %T); these are "+
				"human decision-rate events (select/hover/dragStart/dragEnd/tilt) and cannot "+
				"outrun a geometry loop that runs every real tick — so either this node's "+
				"geometry goroutine has stopped running, or a continuous per-pointer-move "+
				"quantity is being sent here instead of onto a coalescing slot",
			msg.NodeID, InboxDepth, msg.Body))
	}
}

func (n *Messaging) TryRecvExternal() (Msg, bool) {
	select {
	case msg := <-n.extIn:
		return msg, true
	default:
		return Msg{}, false
	}
}

func (n *Messaging) EnqueueSend(destID string, msg Msg) {
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
