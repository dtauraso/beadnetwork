package NodePhiTheta3

import (
	"fmt"
	"strconv"

	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (n *NodePhiTheta3) startVec(j int) polar.Polar {
	if n.Card.K[j-1] == 0 {
		return polar.Polar{}
	}
	a := n.shownValue(j)
	c := n.geom.Constants()
	p := polarindex.OffsetToPolar(polarindex.Offset{Phi: n.spokes(a.Phi, c.MaxIndexPhi), Theta: n.spokes(a.Theta, c.MaxIndexTheta)}, c)
	p.R = float64(n.Card.Start[j-1].R) * n.nodeR
	return p
}

func (n *NodePhiTheta3) start(j int) Angles {
	v, t := n.Card.Start[j-1], n.Card.Tick[j-1]
	return scale(n.Card.K[j-1], Angles{Phi: v.Phi*n.S + t.Phi, Theta: v.Theta*n.S + t.Theta})
}

func (n *NodePhiTheta3) spokes(ticks, wholeTurn int) int {
	perTurn := poleHigh * n.S
	if wholeTurn%perTurn != 0 {
		panic(fmt.Sprintf("NodePhiTheta3.spokes: node %d's scene turn is %d index steps, which is not a whole number of steps per tick at 12s = %d — "+
			"CardPanel.SFits gates s where it is loaded (Startup.NewFromSpec) and where the panel sets it (Dispatch.applyCardKey)",
			n.Me, wholeTurn, perTurn))
	}
	return ticks * (wholeTurn / perTurn)
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(poleHigh * n.S))
}

func (n *NodePhiTheta3) sentVectors() (starts []polar.Polar, leads []NodeCat.Lead) {
	starts = make([]polar.Polar, 0, len(n.Partners))
	for _, j := range n.Partners {
		vec := n.startVec(j)
		starts = append(starts, vec)
		if n.Card.K[j-1] == 1 {
			leads = append(leads, NodeCat.Lead{TargetID: strconv.Itoa(j), Vec: vec})
		}
	}
	return starts, leads
}

func (n *NodePhiTheta3) postArrows() {
	starts, _ := n.sentVectors()
	n.geom.KindPosts().PostStartVectors(starts)
}

func (n *NodePhiTheta3) postVectors() {
	starts, leads := n.sentVectors()
	n.geom.KindPosts().PostStartVectors(starts)
	n.geom.KindPosts().PostLeads(leads)
}

const walkHead = 1

func (n *NodePhiTheta3) placeChain() {
	if n.Me == walkHead {
		n.postVectors()
		return
	}
	starts, leads := n.sentVectors()
	n.geom.KindPosts().PostStartVectors(starts)
	n.geom.KindPosts().PostLeadsHeld(leads)
}
