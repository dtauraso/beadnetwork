package owners

import (
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Node/movemsg"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

type neighborSlot struct {
	ch chan movemsg.Msg
}

func newNeighborSlot() *neighborSlot {
	return &neighborSlot{ch: make(chan movemsg.Msg, 1)}
}

func (s *neighborSlot) deposit(msg movemsg.Msg) {
	mv, ok := msg.Body.(movemsg.Movement)
	if !ok {
		panic(fmt.Sprintf("owners.neighborSlot: a coalescing slot carries only a movemsg.Movement — "+
			"a neighbour's NeighborMoved or the gesture FSM's Drag; got %T. A new body must say how "+
			"two of them merge, by implementing Movement, before it can ride a slot.", msg.Body))
	}
	if mv.Where() == nil && mv.HowFar() == nil {
		panic(fmt.Sprintf("owners.neighborSlot: a coalescing slot was handed a %T naming neither a "+
			"WHERE nor a HOW FAR, so two of them have no defined merge — a WHERE collapses by keeping "+
			"the newest, a HOW FAR by summing, and this is neither.", msg.Body))
	}

	select {
	case unread := <-s.ch:
		if prev, ok := unread.Body.(movemsg.Movement); ok {
			if mv.Where() == nil && prev.HowFar() != nil && mv.HowFar() != nil {
				msg.Body = mv.WithHowFar(polarindex.Sum(*prev.HowFar(), *mv.HowFar()))
			}
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
