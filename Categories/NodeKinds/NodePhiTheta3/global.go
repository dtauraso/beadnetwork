package NodePhiTheta3

import (
	"fmt"
	"strconv"

	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (n *NodePhiTheta3) startVec(v Vec) polar.Polar {
	c := n.geom.Constants()
	p := polarindex.OffsetToPolar(polarindex.Offset{Phi: n.spokes(v.Phi, c.MaxIndexPhi), Theta: n.spokes(v.Theta, c.MaxIndexTheta)}, c)
	p.R = float64(v.R) * n.nodeR
	return p
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

func (n *NodePhiTheta3) postStarts() {
	starts := make([]polar.Polar, 0, len(n.Partners))
	var leads []NodeCat.Lead
	for _, j := range n.Partners {
		vec := n.startVec(scale(n.Card.K[j-1], n.Card.Start[j-1]))
		starts = append(starts, vec)
		if n.Card.K[j-1] == 1 {
			leads = append(leads, NodeCat.Lead{TargetID: strconv.Itoa(j), Vec: vec})
		}
	}
	n.geom.KindPosts().PostStartVectors(starts)
	n.geom.KindPosts().PostLeads(leads)
}
