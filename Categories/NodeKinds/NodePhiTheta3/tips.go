package NodePhiTheta3

import (
	"strconv"

	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

const tipDepth = 16

func (n *NodePhiTheta3) wireTips() {
	me := n.Me - 1
	out := make(map[string]chan<- NodeCat.Tip, len(n.Partners))
	in := make([]<-chan NodeCat.Tip, 0, len(n.Partners))
	for _, j := range n.Partners {
		out[strconv.Itoa(j)] = n.mesh.tip[me][j-1]
		in = append(in, n.mesh.tip[j-1][me])
	}
	n.geom.WireTips(out, in)
}

func (n *NodePhiTheta3) placePartners() {
	var tips []NodeCat.TipPost
	for _, j := range n.Partners {
		if v := n.startVec(j); n.Card.K[j-1] == 1 && v != (polar.Polar{}) {
			tips = append(tips, NodeCat.TipPost{TargetID: strconv.Itoa(j), Vec: v})
		}
	}
	if len(tips) > 0 {
		n.geom.KindPosts().PostTips(tips)
	}
}
