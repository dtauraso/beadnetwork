package clock

import (
	"sync"
	"time"
)

type TickBroadcaster struct {
	register chan chan struct{}
}

func startTickBroadcaster() *TickBroadcaster {
	tb := &TickBroadcaster{register: make(chan chan struct{})}
	go tb.run()
	return tb
}

func (tb *TickBroadcaster) run() {
	ticker := time.NewTicker(tickPeriod)
	defer ticker.Stop()
	var subs []chan struct{}
	for {
		select {
		case ch := <-tb.register:
			subs = append(subs, ch)
		case <-ticker.C:
			for _, ch := range subs {

				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}
}

func (tb *TickBroadcaster) Subscribe() <-chan struct{} {
	pulseCh := make(chan struct{}, 1)
	tb.register <- pulseCh
	return pulseCh
}

var (
	tickBroadcasterOnce sync.Once
	tickBroadcasterInst *TickBroadcaster
)

func globalTickBroadcaster() *TickBroadcaster {
	tickBroadcasterOnce.Do(func() {
		tickBroadcasterInst = startTickBroadcaster()
	})
	return tickBroadcasterInst
}

func NewTickChan() <-chan struct{} {
	return globalTickBroadcaster().Subscribe()
}
