package interior

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/rowevent"
)

const MailboxDepth = 64

type Snapshot struct {
	EventsOnly bool

	Present    []uint8
	Value      []int32
	Ox, Oy, Oz []float32

	Events []rowevent.RowEvent
}

type Mailbox struct {
	ch      chan Snapshot
	nodeRow int32
}

func NewMailbox(nodeRow int32) *Mailbox {
	return &Mailbox{ch: make(chan Snapshot, MailboxDepth), nodeRow: nodeRow}
}

func (m *Mailbox) Send(s Snapshot) {
	if m == nil {
		return
	}
	select {
	case m.ch <- s:
	default:
		panic(fmt.Sprintf(
			"interior.Mailbox.Send: node row %d interior mailbox is full (depth %d) — the node's own "+
				"Update loop is the sole drainer of this mailbox and is not keeping up (missing or "+
				"stalled owners.Interior.WriteFrames call), not a reason to block the sender or drop the "+
				"snapshot",
			m.nodeRow, MailboxDepth))
	}
}

func (m *Mailbox) TryRecv() (Snapshot, bool) {
	if m == nil {
		return Snapshot{}, false
	}
	select {
	case s := <-m.ch:
		return s, true
	default:
		return Snapshot{}, false
	}
}
