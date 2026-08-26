package NodePhi

func (r *restCounters) countArrival() {
	r.msgsSinceOpen++
	r.roundsSinceOpen++
}

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
