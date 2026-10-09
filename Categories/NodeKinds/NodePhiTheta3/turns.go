package NodePhiTheta3

import (
	"strconv"

	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

const turnDepth = 16

var turnCh [nodeCount][nodeCount]chan NodeCat.Turn

func init() {
	for from := 0; from < nodeCount; from++ {
		for to := 0; to < nodeCount; to++ {
			if from != to {
				turnCh[from][to] = make(chan NodeCat.Turn, turnDepth)
			}
		}
	}
}

func (n *NodePhiTheta3) wireTurns() {
	me := n.Me - 1
	out := make(map[string]chan<- NodeCat.Turn, len(n.Partners))
	in := make([]<-chan NodeCat.Turn, 0, len(n.Partners))
	for _, j := range n.Partners {
		out[strconv.Itoa(j)] = turnCh[me][j-1]
		in = append(in, turnCh[j-1][me])
	}
	n.geom.WireTurns(out, in)
}

func (n *NodePhiTheta3) radians(a Angles) (phi, theta float64) {
	c := n.geom.Constants()
	p := polarindex.OffsetToPolar(polarindex.Offset{Phi: n.spokes(a.Phi, c.MaxIndexPhi), Theta: n.spokes(a.Theta, c.MaxIndexTheta)}, c)
	return p.Phi, p.Theta
}

func (n *NodePhiTheta3) turn(j int, change Angles) NodeCat.TurnPost {
	aphi, atheta := n.radians(change)
	return NodeCat.TurnPost{TargetID: strconv.Itoa(j), APhi: aphi, ATheta: atheta}
}

func (n *NodePhiTheta3) postTurns(changes map[int]Angles) {
	var turns []NodeCat.TurnPost
	for _, j := range n.Partners {
		if c, ok := changes[j]; ok && n.Card.K[j-1] == 1 && c != (Angles{}) {
			turns = append(turns, n.turn(j, c))
		}
	}
	if len(turns) > 0 {
		n.geom.KindPosts().PostTurns(turns)
	}
}
