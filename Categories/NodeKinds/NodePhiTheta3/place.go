package NodePhiTheta3

import (
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func (n *NodePhiTheta3) angle(ticks int) float64 {
	return float64(ticks) * 2 * math.Pi / float64(CardPanel.TurnTicks*n.S)
}

func (n *NodePhiTheta3) vector() polar.Polar {
	return polar.Polar{R: float64(n.Card.Start.R) * n.nodeR, Phi: n.angle(n.cur.Phi), Theta: n.angle(n.cur.Theta)}
}

func (n *NodePhiTheta3) postVectors() {
	var out []TiltVectors.StartVector
	for i, j := range n.Partners {
		if n.Card.L[j-1] == 1 {
			out = append(out, TiltVectors.StartVector{ToRow: n.partnerRows[i], Vec: n.vector()})
		}
	}
	n.geom.KindPosts().PostStartVectors(out)
	n.breadcrumb("vectors", fmt.Sprintf("links=%v vectors=%+v", n.Card.L, out))
}

func (n *NodePhiTheta3) postTicks() {
	n.geom.KindPosts().PostTicks(int32(CardPanel.TurnTicks * n.S))
}
