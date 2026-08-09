package PairNode

// rest.go — THE COUNTING THIS NODE DOES ABOUT ITS OWN EXCHANGE, and nothing else.
//
// The counters themselves are restCounters (node_parts.go); these are the two moments they
// change: an arrival counted as it comes in, and the report made after the reply for that
// arrival has gone out. No part of the pair rule reads any of it — a number here cannot move
// a tilt, which is exactly why it is kept away from node.go's decisions.

// countArrival records one received vector: one message and, since this node answers every
// arrival, one round. Counted on the way IN, before the reset marker is even looked at, so a
// message this node genuinely received is counted whatever it turns out to say.
func (r *restCounters) countArrival() {
	r.msgsSinceOpen++ // this node's own receive
	r.roundsSinceOpen++
}

// reportRest reports this node's counts to its own geometry, AFTER the reply for this cycle
// has been sent, and freezes them the first time the rule found itself already at rest.
//
// LIVE while the tilt is still turning: report the running counts each round so the
// readout climbs as it goes rather than staying blank and then jumping. Once this
// node comes to rest the same numbers are reported one last time and then frozen —
// the exchange keeps circulating after rest (handleVectorCycle replies to every arrival
// whether or not it moved), so a counter that kept reporting would measure how long
// the scene had been open instead of how far the tilt travelled.
func (n *Node) reportRest() {
	if !n.rest.restReported {
		n.rest.roundsAtRest = n.rest.roundsSinceOpen
		n.rest.msgsAtRest = n.rest.msgsSinceOpen
		if n.plumb.Self != nil {
			n.plumb.Self.SetRoundsToParallel(n.rest.roundsAtRest, n.rest.msgsAtRest)
		}
		if n.rest.restedThisCycle {
			n.rest.restReported = true
		}
	}
	n.rest.restedThisCycle = false
}
