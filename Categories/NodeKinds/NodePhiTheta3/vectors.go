package NodePhiTheta3

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/AngleArcs"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func (n *NodePhiTheta3) postVectors() {
	var out []polar.Polar
	vecs, picked := n.peer.OutVecs()
	for _, v := range vecs {
		out = append(out, polar.Cart2polar(v))
	}
	n.geom.KindPosts().PostStartVectors(out, picked)
	n.geom.KindPosts().PostShownEdges(n.peer.PickedOuts())
	n.postArcs()
}

func (n *NodePhiTheta3) postArcs() {
	s := n.peer.S
	var arcs []AngleArcs.Angle
	for _, l := range n.peer.Links() {
		if !l.Picked || l.Dir(s).R*n.peer.Frame.Rho < 1e-9 {
			continue
		}
		phi, theta := n.peer.Frame.Arcs(s, l)
		arcs = append(arcs, AngleArcs.Angle{At: n.peer.Offset(l.From), Phi: AngleArcs.Path(phi), Theta: AngleArcs.Path(theta)})
	}
	n.geom.KindPosts().PostArcs(arcs)
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(Links.TurnTicks * n.peer.S))
}
