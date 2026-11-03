package NodePhiTheta3

import (
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
	var arcs []AngleArcs.Angle
	for _, l := range n.peer.Links() {
		if l.Dir(s).R*n.peer.Frame.Rho < 1e-9 {
			continue
		}
		phi, theta := n.peer.Frame.Arcs(s, l)
		arcs = append(arcs, AngleArcs.Angle{At: n.peer.Offset(l.From), Phi: phi, Theta: theta})
	}
	n.geom.KindPosts().PostArcs(arcs)
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(Links.TurnTicks * n.peer.S))
}
