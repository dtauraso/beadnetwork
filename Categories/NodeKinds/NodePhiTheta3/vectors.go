package NodePhiTheta3

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Node/AngleArcs"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func (n *NodePhiTheta3) postVectors() {
	var out []polar.Polar
	for _, v := range n.peer.Outs() {
		out = append(out, polar.Cart2polar(v))
	}
	n.geom.KindPosts().PostStartVectors(out)
	n.postArcs()
}

func (n *NodePhiTheta3) postArcs() {
	s := n.peer.S
	tick := 2 * math.Pi / float64(Links.TurnTicks*s)
	var arcs []AngleArcs.Angle
	for _, l := range n.peer.Links() {
		d := l.Dir(s)
		arcs = append(arcs, AngleArcs.Angle{
			At:  n.peer.Offset(l.From),
			Dir: polar.Polar{R: d.R * n.peer.Frame.Rho, Phi: d.Phi * tick, Theta: d.Theta * tick},
		})
	}
	n.geom.KindPosts().PostArcs(arcs)
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(Links.TurnTicks * n.peer.S))
}
