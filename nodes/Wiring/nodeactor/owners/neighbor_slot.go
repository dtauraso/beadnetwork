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
	if msg.Kind != movemsg.KindCenter && msg.Kind != movemsg.KindDrag {
		panic("owners.neighborSlot: a coalescing slot carries only KindCenter from a neighbour or " +
			"KindDrag from the gesture FSM; got kind " + msg.Kind +
			". A new message kind must say how two of them merge before it can ride a slot.")
	}
	if msg.Target == nil && msg.Delta == nil {
		panic("owners.neighborSlot: a coalescing slot was handed a message carrying neither a Target " +
			"nor a Delta, so two of them have no defined merge — a WHERE collapses by keeping the " +
			"newest, a HOW FAR by summing, and this is neither; kind " + msg.Kind)
	}

	select {
	case unread := <-s.ch:
		if msg.Target == nil && unread.Delta != nil && msg.Delta != nil {
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
