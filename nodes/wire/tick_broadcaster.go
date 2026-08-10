// tick_broadcaster.go — TickBroadcaster, the ONE goroutine in the process that
// ever waits on wall time. Every RealClock (real_clock.go) subscribes against
// the single instance here and paces itself off the pulse (sleep_cycle.go).

package wire

import (
	"sync"
	"time"
)

// TickBroadcaster is the ONE thing in the process that waits on wall time. Its
// single goroutine (run) owns a time.Ticker and, on every fire, pushes a pulse
// to every subscriber's own dedicated channel — non-blockingly, so a slow
// subscriber never stalls the broadcaster or any other subscriber. Every other
// goroutine in the network (movers, node loops, DriveHeld, gate loops) blocks
// on RECEIVE from its own subscription instead of sleeping — see clock.go's
// package doc and PLAN.md "No sleeping".
type TickBroadcaster struct {
	// register is how a new subscriber hands the broadcaster its channel to add
	// to the fan-out set. Unbuffered: the broadcaster's run loop is always
	// selecting on it (or the ticker), so a send here never blocks past one
	// broadcaster loop iteration.
	register chan chan struct{}
}

// startTickBroadcaster starts the broadcaster's one goroutine and returns a
// handle new subscribers register against. It runs for the life of the
// process — there is exactly one of these (see globalTickBroadcaster), no
// per-goroutine tickers anywhere else.
func startTickBroadcaster() *TickBroadcaster {
	tb := &TickBroadcaster{register: make(chan chan struct{})}
	go tb.run()
	return tb
}

// run is the ONE goroutine in the process that ever calls time.NewTicker (or
// blocks on it). It owns the fan-out subscriber list — nothing else reads or
// writes subs, so it needs no lock.
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
				// Non-blocking, coalescing send: a subscriber that has not yet
				// drained the previous pulse just sees "at least one tick has
				// fired since I last checked" rather than backing up a queue —
				// the same latest-wins shape as SendSpeedNonBlocking/
				// SendLatestNonBlocking in speed_delivery.go.
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}
}

// Subscribe returns a fresh, dedicated buffered-1 channel that receives a
// pulse once per wall tick from this broadcaster. Call once per goroutine (or
// once per Clock value, which is the same thing under per-goroutine-clock.md);
// the returned channel is owned by the caller from then on and must not be
// shared with a second goroutine.
func (tb *TickBroadcaster) Subscribe() <-chan struct{} {
	pulseCh := make(chan struct{}, 1) // chan-name-ok: internal broadcaster->subscriber pulse, not a node-to-node wire
	tb.register <- pulseCh
	return pulseCh
}

var (
	tickBroadcasterOnce sync.Once
	tickBroadcasterInst *TickBroadcaster
)

// globalTickBroadcaster lazily starts (sync.Once — coordinates goroutine
// startup, not shared mutable state, so it stays inside check-no-network-locks'
// allowance for sync.Once) the process-wide single clock goroutine on first
// use and returns it thereafter. Every RealClock (NewRealClock and Copy)
// subscribes against this one instance, so however many clock-holders the
// process has, exactly one goroutine ever waits on wall time.
func globalTickBroadcaster() *TickBroadcaster {
	tickBroadcasterOnce.Do(func() {
		tickBroadcasterInst = startTickBroadcaster()
	})
	return tickBroadcasterInst
}

// NewTickChan returns a fresh dedicated tick-pulse channel from the global
// broadcaster, for callers that pace themselves without going through a full
// Clock (e.g. gatecommon's no-loader wall-clock fallbacks). Call once per
// goroutine, same rule as Subscribe.
func NewTickChan() <-chan struct{} {
	return globalTickBroadcaster().Subscribe()
}
