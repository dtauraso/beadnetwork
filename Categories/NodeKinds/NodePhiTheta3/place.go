package NodePhiTheta3

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func (n *NodePhiTheta3) angle(ticks int) float64 {
	return float64(ticks) * 2 * math.Pi / float64(CardPanel.TurnTicks*n.S)
}

func (n *NodePhiTheta3) vectorOf(p say) polar.Polar {
	return polar.Polar{R: float64(p.R) * n.nodeR, Phi: n.angle(p.A.Phi), Theta: n.angle(p.A.Theta)}
}

func (n *NodePhiTheta3) otherPartner(j int) int {
	if n.Partners[0] == j {
		return n.Partners[1]
	}
	return n.Partners[0]
}

func (n *NodePhiTheta3) tips() []NodeCat.TipPost {
	var out []NodeCat.TipPost
	for _, j := range n.Partners {
		if n.Card.L[j-1] != 1 || !n.known[j-1] {
			continue
		}
		tp := NodeCat.TipPost{TargetID: strconv.Itoa(j), Vec: n.vectorOf(n.peer[j-1])}
		if g := n.otherPartner(j); n.known[g-1] && n.peer[g-1].L[j-1] == 1 {
			tp.With = strconv.Itoa(g)
		}
		out = append(out, tp)
	}
	return out
}

func (n *NodePhiTheta3) place(force bool) {
	tips := n.tips()
	if !force && slices.Equal(tips, n.posted) {
		return
	}
	n.posted = tips
	n.geom.KindPosts().PostTips(tips)
	n.breadcrumb("place", fmt.Sprintf("tips=%+v", tips))
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(CardPanel.TurnTicks * n.S))
}
