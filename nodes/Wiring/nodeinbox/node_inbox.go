package nodeinbox

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

type NodeInboxes struct {
	tiltEdit map[string]chan movemsg.TiltEditMsg

	lattice map[string]chan int32
}

func (ib *NodeInboxes) ClaimLatticeIn(id string, ch chan int32) {
	if ib.lattice == nil {
		ib.lattice = map[string]chan int32{}
	}
	ib.lattice[id] = ch
}

func (ib *NodeInboxes) ClaimTiltEditIn(id string, ch chan movemsg.TiltEditMsg) {
	if ib.tiltEdit == nil {
		ib.tiltEdit = map[string]chan movemsg.TiltEditMsg{}
	}
	ib.tiltEdit[id] = ch
}

func (ib *NodeInboxes) BroadcastLatticePoints(points int32) {
	for _, ch := range ib.lattice {
		scenepersist.SendLatticePointsNonBlocking(ch, points)
	}
}

func (ib *NodeInboxes) SendTiltEdit(ctx context.Context, id string, msg movemsg.TiltEditMsg) bool {
	ch, ok := ib.tiltEdit[id]
	if !ok {
		return false
	}
	if ctx == nil {
		ch <- msg
		return true
	}
	select {
	case ch <- msg:
	case <-ctx.Done():
	}
	return true
}
