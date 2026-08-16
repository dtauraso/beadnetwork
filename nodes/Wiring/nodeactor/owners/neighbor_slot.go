package owners

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

type neighborSlot struct {
	ch chan movemsg.Msg
}

func newNeighborSlot() *neighborSlot {
	return &neighborSlot{ch: make(chan movemsg.Msg, 1)}
}

func (s *neighborSlot) deposit(msg movemsg.Msg) {
	if msg.Delta == nil || (msg.Kind != movemsg.KindCenter && msg.Kind != movemsg.KindDrag) {
		panic("owners.neighborSlot: a coalescing slot carries only an INCREMENTAL DELTA — " +
			"KindCenter from a neighbour, or KindDrag from the gesture FSM — because summing " +
			"is the only way two of them may collapse into one; got kind " + msg.Kind +
			". A new message kind must say how two of them merge before it can ride a slot.")
	}
	select {
	case unread := <-s.ch:
		if unread.Delta != nil {
			summed := polarindex.Sum(*unread.Delta, *msg.Delta)
			msg.Delta = &summed
		}
	default:
	}
	s.ch <- msg
}

func (s *neighborSlot) take() (movemsg.Msg, bool) {
	select {
	case msg := <-s.ch:
		return msg, true
	default:
		return movemsg.Msg{}, false
	}
}
