package NodePhiTheta3

import (
	"math"
	"strconv"

	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (n *NodePhiTheta3) offsetOf(v Vec) polarindex.Offset {
	c := n.geom.Constants()
	r := int(math.Round(float64(v.R) * n.stepsPerR))
	return polarindex.Offset{Phi: n.spokes(v.Phi, c.MaxIndexPhi), Theta: n.spokes(v.Theta, c.MaxIndexTheta), R: r}
}

func (n *NodePhiTheta3) spokes(ticks, wholeTurn int) int {
	return int(math.Round(float64(ticks*wholeTurn) / float64(poleHigh*n.S)))
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(poleHigh * n.S))
}

func (n *NodePhiTheta3) postStarts() {
	starts := make([]polarindex.Offset, 0, len(n.Partners))
	var leads []NodeCat.Lead
	for _, j := range n.Partners {
		vec := n.offsetOf(n.Card.Start[j-1])
		starts = append(starts, vec)
		if n.Card.K[j-1] == 1 {
			leads = append(leads, NodeCat.Lead{TargetID: strconv.Itoa(j), Vec: vec})
		}
	}
	n.geom.KindPosts().PostStartVectors(starts)
	if len(leads) > 0 {
		n.geom.KindPosts().PostLeads(leads)
	}
}
