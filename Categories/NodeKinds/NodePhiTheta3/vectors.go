package NodePhiTheta3

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func (n *NodePhiTheta3) postVectors() {
	s := n.peer.S
	tick := 2 * math.Pi / float64(Links.TurnTicks*s)
	var out []polar.Polar
	for _, l := range n.peer.Links() {
		d := l.Dir(s)
		out = append(out, polar.Polar{R: d.R * n.peer.Frame.Rho, Phi: d.Phi * tick, Theta: d.Theta * tick})
	}
	n.geom.KindPosts().PostStartVectors(out)
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(Links.TurnTicks * n.peer.S))
}
