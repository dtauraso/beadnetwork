package NodePhiTheta3

import (
	"strconv"

	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

const placementDepth = 16

var place [nodeCount][nodeCount]chan NodeCat.Placement

func init() {
	for from := 0; from < nodeCount; from++ {
		for to := 0; to < nodeCount; to++ {
			if from != to {
				place[from][to] = make(chan NodeCat.Placement, placementDepth)
			}
		}
	}
}

func (n *NodePhiTheta3) wirePlacement() {
	me := n.Me - 1
	out := make(map[string]chan<- NodeCat.Placement, len(n.Partners))
	in := make([]<-chan NodeCat.Placement, 0, len(n.Partners))
	for _, j := range n.Partners {
		out[strconv.Itoa(j)] = place[me][j-1]
		in = append(in, place[j-1][me])
	}
	n.geom.WirePlacement(out, in)
}
