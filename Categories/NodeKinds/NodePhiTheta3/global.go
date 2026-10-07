package NodePhiTheta3

import (
	"math"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (n *NodePhiTheta3) offsetOf(v Vec) polarindex.Offset {
	r := int(math.Round(float64(v.R) * n.stepsPerR))
	return polarindex.Offset{Phi: v.Phi * n.S, Theta: v.Theta * n.S, R: r}
}

func (n *NodePhiTheta3) parent() (int, bool) {
	a, b := n.Partners[0], n.Partners[1]
	switch {
	case n.Card.K[a-1] == 1 && n.Card.K[b-1] == 0:
		return a, true
	case n.Card.K[b-1] == 1 && n.Card.K[a-1] == 0:
		return b, true
	}
	return 0, false
}

func (n *NodePhiTheta3) place() {
	j, ok := n.parent()
	if !ok {
		return
	}
	n.geom.KindPosts().PostVectorFrom(strconv.Itoa(j), n.offsetOf(n.arrival[j-1]))
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(poleHigh * n.S))
}

func (n *NodePhiTheta3) postStarts() {
	starts := make([]polarindex.Offset, 0, len(n.Partners))
	for _, j := range n.Partners {
		starts = append(starts, n.offsetOf(n.Card.Start[j-1]))
	}
	n.geom.KindPosts().PostStartVectors(starts)
}
