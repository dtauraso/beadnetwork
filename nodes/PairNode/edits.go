package PairNode

// edits.go — WHAT A USER DOES TO THIS NODE FROM THE PANEL, and what a reset undoes.
//
// The three panel-driven edits arrive on one channel (TiltEditIn) and are split apart by
// applyTiltEdit: a ▲/▼ click, START, and RESET. RESET is the one that does more than move an
// index, so clear and the drain it needs live here beside it — a reset is defined by what it
// leaves behind, which is nothing, and that is a statement about this node's channels and
// beads rather than about the pair rule.
//
// Nothing here decides where a tilt comes to rest. The rule is machine.go's and the per-cycle
// decision is node.go's; this file only starts and stops the exchange those run.

import (
	"github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// drainTiltEdit drains TiltEditIn non-blocking: a panel/RESET/START edit — see the package doc
// comment for the three-way split. applyTiltEdit decides placeBead: true for Start
// and for a plain adjust (both open the exchange), false only for Reset.
func (n *Node) drainTiltEdit(clk clock.Clock) {
	if n.tilt.TiltEditIn == nil {
		return
	}
	select {
	case edit := <-n.tilt.TiltEditIn:
		placeBead := n.applyTiltEdit(edit)
		n.syncTiltIndex()
		if placeBead && n.plumb.Out != nil {
			n.plumb.Out.PlaceDrivenAt(1, clk.Tick())
		}
	default:
	}
}

// applyTiltEdit applies one panel-driven edit — TiltVectorAnglePanel's ±1 click, the START
// TILT button (TiltVectorButtons.tsx), or the RESET button (same file) — directly to this
// node's OWN indices, same no-mover-round-trip shape as stepFromVector. Reports whether the
// caller should place a bead: true for Start and for a plain adjust, false only for a reset.
//
// The three branches are now split (task/pair-node-owns-itself):
//
//   - Reset: a stop-and-return, not a nudge — see clear's own doc comment.
//   - Start: opens the vector exchange from whatever angles are CURRENTLY set, by sending
//     this node's own outgoingVector on VectorOut and placing a bead. It changes NO index.
//     This is the vector channel's whole starting move, and without it the channel never
//     carries a direction at all: handleVectorCycle only ever sends in REPLY to an arrival,
//     so with nothing to reply to, no node ever received one, no node ever set its
//     received-direction record, and the third arrow could not be drawn anywhere.
//   - a plain adjust (neither Reset nor Start): applies the ±1 click to the named axis, marks
//     this end Held (a tilt a user set is intent, so this end keeps its index and does not
//     turn on an arrival — the partner moves instead), and ALSO OPENS THE EXCHANGE by sending
//     this node's own outgoing vector and asking the caller to place a bead. Previously an
//     adjust sent nothing, which is why a tilted pair just sat there: the tilted end went deaf
//     and silent at once.
func (n *Node) applyTiltEdit(edit Wiring.TiltEditMsg) (placeBead bool) {
	if edit.Reset {
		n.clear()
		// Tell the partner, so it clears too — see clear's own doc comment for why the
		// partner's clear, not this one, is what actually ends the exchange.
		Wiring.SendVectorLatestNonBlocking(n.vec.VectorOut, Wiring.TiltVectorMsg{Reset: true})
		return false
	}
	if edit.Start {
		// START BELONGS TO PAIR ID 1 ALONE. The exchange is begun from ONE end, so there is
		// exactly one opening direction to answer; opened from both, each end also replies to
		// the other's opener in the same round — two exchanges running through one pair of
		// channels rather than the one a user asked for, which shows up as the pair settling
		// and then being kicked off its rest state again, forever.
		//
		// The panel sends START to every node it lists (TiltVectorButtons.tsx posts one
		// record per row, exactly as RESET does), because the WEBVIEW must not know which end
		// is which — that is domain knowledge, and TS holds none. Go decides, here, by id.
		if n.plumb.PairID != 1 {
			return false
		}
		// Open the vector exchange from the current angles — see this function's own doc
		// comment. Sends exactly what the old adjust-side-effect kick sent, but changes no
		// index of its own.
		// The opener is deliberately NOT counted. It is the kick that starts the exchange,
		// not part of a round, and counting it would make the opening end report one more
		// message than the other for the same amount of work — the two ends did the same
		// number of rounds and the same number of receive/reply pairs.
		Wiring.SendVectorLatestNonBlocking(n.vec.VectorOut, n.outgoingVector())
		return true
	}
	// A click names the TOP — it is the arrow the user is dragging, not a measured end.
	if edit.Up {
		n.setTop(n.topState().next)
	} else {
		n.setTop(n.topState().prev)
	}
	// AND IT STOPS THERE: no send, no bead, and NO MACHINE CHOSEN. A click is not the moment
	// to read the gap — the setup is not finished until START, and a user clicks their way up
	// to the angle they want. Deciding on the first click read a gap of one step and locked
	// the pair to the parallel machine while the tilt was still eleven clicks from where it
	// was going. The choice is made when the exchange opens (handleVectorCycle).
	// AND IT STOPS THERE: no send, no bead. Setting an angle and running the exchange are
	// separate acts — a click moves this node's own tilt and nothing else happens until START.
	return false
}

// clear returns THIS node to its opening state and — the part that matters — leaves
// nothing behind that could restart the straightening exchange. A reset is not "set the
// indices to 0"; it is "there is no message anywhere in the pair", and zeroed indices are
// just what that looks like from outside. Everything the pair holds between clicks is
// cleared here, each piece by the goroutine that owns it:
//
//   - this node's own tilt and derived coplanar-normal indices (owned here);
//   - this node's record of the last received direction, the third drawn arrow (owned here);
//   - this node's VectorIn, drained non-blocking — the receive end is owned here, and a
//     direction already sitting in it would arrive on the next cycle and step the tilt
//     straight back off zero. Depth-1 latest-wins, so one receive empties it;
//   - this node's already-DELIVERED beads, drained off In the same way and for the same
//     reason — the bead edge paces the exchange that turns
//     these tilts (the bead paces each round trip of the vector exchange), so a reset that skips it
//     visibly does not take;
//   - this node's OUTGOING beads, still crossing. A PacedWire is driven by its source
//     node's own mover state, which for this kind is this goroutine's own Self, so it drops
//     them through that (ClearOutBeads / PairNodeSelf.ClearOutBeads) rather than reaching
//     into the wire.
//
// WHY BOTH CALLERS EXIST. The RESET button sends one record per node, but the two nodes
// act on their own goroutines at their own moments, so a single clear each is racy: the
// partner can place one more bead in the window after this node cleared, and it lands
// afterwards and restarts everything. What closes that window is the Reset MARKER — this
// node clears, sends the marker, and places nothing ever again from that path; the partner
// runs this same clear when the marker arrives, which is therefore ordered after the last
// thing this node could have placed. So each node clears twice, and the second one is the
// one that provably lands last. The marker gets no reply (handleVectorCycle), so it stops
// there instead of bouncing.
func (n *Node) clear() {
	n.setTop(n.ringOf().at(0))
	// The machine this node was running goes too — RESET is the one thing that releases it, and
	// what it returns to is the setting mode, which is also where a fresh node starts.
	n.tilt.Machine = setting
	n.syncTiltIndex()
	n.vec.ReceivedThetaIdx = 0
	n.vec.ReceivedSet = false
	n.syncReceivedVector()
	// The counters go with the machine: RESET returns this node to the setting mode, so the
	// next START opens a fresh exchange and its rounds are counted from zero, not continued.
	n.rest = restCounters{}
	if n.plumb.Self != nil {
		n.plumb.Self.SetRoundsToParallel(0, 0)
	}
	Wiring.PollRecvVector(n.vec.VectorIn)
	n.drainIn()
	if n.plumb.ClearOutBeads != nil {
		n.plumb.ClearOutBeads()
	}
}

// drainIn empties this node's own In of every bead already delivered to it, on this
// node's own goroutine (In.PollRecv is non-blocking, so this terminates as soon as the
// queue is empty). Bounded by what the partner placed before it stopped placing, which
// is the pair's own bead-per-cycle traffic — the same drain-until-empty shape as
// PacedWire.drainPlacements, whose doc comment carries the full reasoning for why these
// loops need no cap.
func (n *Node) drainIn() {
	if n.plumb.In == nil {
		return
	}
	for {
		if _, ok := n.plumb.In.PollRecv(); !ok {
			return
		}
	}
}
