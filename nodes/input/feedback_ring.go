// feedback_ring.go — Input's feedback-ring emit path, entered from Update only when
// FeedbackIn.Wired() is true. Grouped together because the three functions here are one
// state machine: updateFeedbackRing is the flat cycle loop, feedbackRingSend is its
// peek+send phase, and feedbackRingReact is its react-to-feedback phase.

package input

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// updateFeedbackRing runs the feedback-ring emit path. It returns when ctx is
// cancelled or FeedbackIn closes. Called only when FeedbackIn.Wired() is true.
//
// Feedback ring: PEEK+SEND then READ. Sending does NOT deplete the buffer —
// each iteration peeks the END of working and launches that bead; the buffer
// stays full (4) at rest. The FIRST send is just the normal loop body
// (peek+send) running before any feedback is read, so the ring self-starts
// with no special seed and no t=0 deadlock.
//
// After sending, READ node 2's feedback s on FeedbackIn:
//
//	s == 1 -> POP the end (the "change the bead" action); refill on empty.
//	s == 0 -> hold: do nothing, keep sending the same last bead next loop.
func (n *Node) updateFeedbackRing(ctx context.Context, working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) {
	// clk is this goroutine's own copy, taken once by the caller (Update) at
	// startup. Do not re-derive from n.clock() here; that would be a second,
	// independent copy from the same shared source, defeating "one copy per
	// goroutine".

	// ONE flat loop, identical in shape to the plain source path below: each
	// cycle does exactly one step of work, with NO nested wait loop. The
	// "waiting for node 2's feedback step" that used to be an inner loop is now
	// the flat `awaiting` flag carried across cycles: when false we peek+send a
	// fresh bead and arm the wait; when true we are mid-traversal and just
	// step+poll. Layout/drag handling is NOT here — it lives in the node's
	// dedicated always-on layout goroutine (split-layout-bead-goroutines.md), so
	// this bead loop is purely the pausable half and dragging is unaffected by
	// whatever this loop is waiting on.
	awaiting := false
	for {
		if ctx.Err() != nil {
			return
		}

		if !awaiting {
			if !n.feedbackRingSend(working, backup, init, emitBeads, clk) {
				return
			}
			awaiting = true
		}

		// One step per cycle: sleep and poll FeedbackIn non-blocking. Each
		// broadcast wire's own goroutine advances its in-flight beads; this node
		// is never parked across the traversal and no longer steps the wires
		// itself.
		clock.ApplySpeedNonBlocking(clk, n.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}

		s, ok := n.FeedbackIn.PollRecv()
		if !ok {
			// Time's step has not arrived yet — keep cycling (drains
			// Layout again next pass). Still awaiting.
			continue
		}
		// Feedback arrived: re-arm the next peek+send regardless of hold/pop.
		awaiting = false
		n.feedbackRingReact(s, working, backup, init, emitBeads, clk)
	}
}

// feedbackRingSend is updateFeedbackRing's peek+send phase: guard the empty
// buffer, PEEK the end (do NOT reslice) and SEND. Buffer unchanged. Input
// places the same bead on every wired output the same cycle (broadcastPlace —
// preserves concurrent broadcast) so whatever is wired to OutCadence and
// whatever is wired to ToExcitatory traverse in lockstep. Returns false only
// on the same terminal failure broadcastPlace itself reports (mirrors its own
// doc — a momentarily full paced-wire buffer is transient and never reaches
// here as false).
func (n *Node) feedbackRingSend(working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) bool {
	// Guard: never peek an empty slice. Refill keeps working non-empty, but be
	// safe.
	if len(*working) == 0 {
		*working = *backup
		*backup = append([]int(nil), init...)
		emitBeads()
	}

	v := (*working)[len(*working)-1]
	if n.Fire != nil {
		n.Fire()
	}
	return n.broadcastPlace(v, clk.Tick())
}

// feedbackRingReact is updateFeedbackRing's react-to-feedback phase, run once
// node 2's feedback s has arrived this cycle:
//
//	s == 1 -> POP the end (the "change the bead" action); refill on empty.
//	s == 0 -> hold: do nothing, keep sending the same last bead next loop.
func (n *Node) feedbackRingReact(s int, working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) {
	if s != 1 {
		// Hold: buffer unchanged, send the same last bead next cycle.
		return
	}

	// s == 1: POP the end (change the bead); refill when working empties.
	*working = (*working)[:len(*working)-1]
	if len(*working) == 0 {
		// Animated refill: the top row (backup) SLIDES DOWN into the working
		// row at human speed (clock-paced, pause-aware). After the slide
		// lands, the new top row appears via the full emitBeads below.
		if n.EmitRefillSlide != nil {
			n.EmitRefillSlide(clk, n.SpeedCh, *backup)
		}
		*working = *backup
		*backup = append([]int(nil), init...)
	}
	emitBeads() // array changed (pop, maybe refill) → restream interior
}
