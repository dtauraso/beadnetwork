// speed_delivery.go — per-goroutine-clock.md "Delivery": the non-blocking,
// latest-wins helpers a goroutine uses to poll for a speed/held-value change on
// its own inbox channel, and the sender-side counterparts.

package wire

// ApplySpeedNonBlocking is the delivery half of per-goroutine-clock.md
// "Delivery": every paced loop grows exactly this one poll, folded into its
// existing sleep/select point. speedCh is a buffered-1, latest-wins channel
// built once at load time (see loader.go / builders.go) and owned from then on
// by exactly the one goroutine that reads it — nothing else may read it.
// A pending value (if any) is drained and applied to clk's OWN
// copy via SetSpeed; an empty channel is a no-op; a nil channel (unwired
// goroutines, or test builds constructed with no loader) is always a no-op
// too, since a receive on a nil channel is never selected. This is
// non-blocking on purpose — a goroutine that is not yet awake must never be
// forced to wake early just to drain its inbox.
func ApplySpeedNonBlocking(clk Clock, speedCh <-chan float64) {
	select {
	case sp := <-speedCh:
		if rc, ok := clk.(*RealClock); ok {
			rc.SetSpeed(sp)
		}
	default:
	}
}

// SendSpeedNonBlocking is the send half of Delivery: it delivers speed to one
// clock-holder's buffered-1 channel WITHOUT blocking on a goroutine that may be
// asleep or never reads. If the buffer already holds a stale pending value
// (a rapid second change arrived before the holder woke to drain the first),
// that stale value is dropped and replaced — LATEST WINS is correct because
// speed is absolute state, not an event stream (per-goroutine-clock.md
// "Delivery"). ch must be a channel this call's caller alone sends on (the
// stdin-reader goroutine, which is the sole writer of every channel collected
// at load) — sending from two goroutines onto the same ch would race the
// drain-then-send pair below.
func SendSpeedNonBlocking(ch chan float64, speed float64) {
	select {
	case ch <- speed:
		return
	default:
	}
	// Buffer full: drain the stale value, then place the new one. Both steps are
	// non-blocking; if some other reader raced us and drained it first between
	// the two selects, the second send below still succeeds (buffer now empty).
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- speed:
	default:
	}
}

// SendLatestNonBlocking is the int64 twin of SendSpeedNonBlocking, used for
// delivering a "held" value from a node's main loop (the owner) to its own
// spawned DriveHeld goroutine(s) over a buffered-1, latest-wins channel — the
// same non-blocking drain-then-send shape, because held (like speed) is
// absolute state, not an event stream: a goroutine that wakes up cares only
// about the CURRENT held value, not every intermediate one it missed while
// asleep. ch must be a channel this call's caller alone sends on (the node's
// own main-loop goroutine, which owns the held value); sending from two
// goroutines onto the same ch would race the drain-then-send pair below. A
// node driving two outputs off one held value (e.g. Pulse's Out/OutFanout) must
// pass a DIFFERENT channel per DriveHeld goroutine and call this once per
// channel — passing the same channel to two DriveHeld goroutines would starve
// whichever one loses a given receive (exactly the speedCh rationale above).
func SendLatestNonBlocking(ch chan int64, v int64) {
	select {
	case ch <- v:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- v:
	default:
	}
}
