package NodePhiTheta3

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func (n *NodePhiTheta3) postVectors() {
	var out []polar.Polar
	for _, v := range n.peer.Outs() {
		out = append(out, polar.Cart2polar(v))
	}
	n.geom.KindPosts().PostStartVectors(out)
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(Links.TurnTicks * n.peer.S))
}
