package NodePhiTheta3

import (
	"strconv"

	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

const tipDepth = 16

type say struct {
	L [nodeCount]int
	R int
	A Angles
}

type mesh struct {
	say  [nodeCount][nodeCount]chan say
	tip  [nodeCount][nodeCount]chan NodeCat.Tip
	wake [nodeCount]chan<- struct{}
}

func newMesh() any {
	m := &mesh{}
	for from := 0; from < nodeCount; from++ {
		for to := 0; to < nodeCount; to++ {
			if from != to {
				m.say[from][to] = make(chan say, 1)
				m.tip[from][to] = make(chan NodeCat.Tip, tipDepth)
			}
		}
	}
	return m
}

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

func (n *NodePhiTheta3) tell() {
	msg := say{L: n.Card.L, R: n.Card.Start.R, A: n.cur}
	for _, j := range n.Partners {
		ch := n.mesh.say[n.Me-1][j-1]
		select {
		case <-ch:
		default:
		}
		ch <- msg
		if w := n.mesh.wake[j-1]; w != nil {
			select {
			case w <- struct{}{}:
			default:
			}
		}
	}
}

func (n *NodePhiTheta3) hear() bool {
	heard := false
	for _, j := range n.Partners {
		select {
		case m := <-n.mesh.say[j-1][n.Me-1]:
			n.peer[j-1], n.known[j-1], heard = m, true, true
		default:
		}
	}
	return heard
}
