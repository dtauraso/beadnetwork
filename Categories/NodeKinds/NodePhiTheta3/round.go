package NodePhiTheta3

import "fmt"

func (n *NodePhiTheta3) round(in [nodeCount]pathMsg) {
	a, b := n.Partners[0], n.Partners[1]
	for _, j := range n.Partners {
		if n.Card.L[j-1] == 0 {
			in[j-1] = pathMsg{}
		}
	}

	chosen := pickOne(in[a-1].K, in[b-1].K, in[a-1].A, in[b-1].A)
	directions := map[int]Angles{}
	for _, j := range n.Partners {
		n.arrival[j-1] = step(chosen, n.Card.PoleOffset[j-1], n.M, n.S)
		directions[j] = sub(n.arrival[j-1], chosen)
	}
	n.postTurns(directions)
	if !n.started {
		for _, j := range n.Partners {
			n.shown[j-1] = n.start(j)
		}
	}
	n.started = true
	for _, j := range n.Partners {
		if v := n.sendValue(j); v != (Angles{}) {
			n.shown[j-1] = v
		}
	}
	n.postArrows()

	if !n.loggedOnce || n.arrival != n.logged {
		n.breadcrumb("card", fmt.Sprintf("in %d=%v %d=%v chosen=%v arrival %d=%v %d=%v",
			a, in[a-1], b, in[b-1], chosen, a, n.arrival[a-1], b, n.arrival[b-1]))
		n.logged, n.loggedOnce = n.arrival, true
	}
}
