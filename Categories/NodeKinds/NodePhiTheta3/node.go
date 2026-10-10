package NodePhiTheta3

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

const nodeCount = CardPanel.NodeCount

type pathMsg struct {
	K int
	A Angles
}

type mesh struct {
	link [nodeCount][nodeCount]chan pathMsg
	tip  [nodeCount][nodeCount]chan NodeCat.Tip
}

func newMesh() any {
	m := &mesh{}
	for from := 0; from < nodeCount; from++ {
		for to := 0; to < nodeCount; to++ {
			if from != to {
				m.link[from][to] = make(chan pathMsg)
				m.tip[from][to] = make(chan NodeCat.Tip, tipDepth)
			}
		}
	}
	return m
}

type NodePhiTheta3 struct {
	geom *NodeCat.NodeGeometry
	mesh *mesh

	Clock   clock.Clock
	SpeedCh <-chan float64
	EditIn  <-chan CardPanel.EditMsg
	StepIn  <-chan struct{}
	ResetIn <-chan struct{}
	StartIn <-chan struct{}
	Wake    <-chan struct{}

	steps int
	reset bool

	Me       int
	Partners [nodeCount - 1]int

	Card CardPanel.Card
	S    int
	M    int

	nodeR float64

	arrival [nodeCount]Angles
	shown   [nodeCount]Angles
	started bool

	logged     [nodeCount]Angles
	loggedOnce bool
}

func (n *NodePhiTheta3) breadcrumb(label, value string) {
	n.geom.Trace().Post([]NodeCat.RowEvent{{
		Kind: NodeCat.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: n.geom.Stream().NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (n *NodePhiTheta3) sendValue(j int) Angles {
	if !n.started {
		return n.start(j)
	}
	return scale(n.Card.K[j-1], n.arrival[j-1])
}

func (n *NodePhiTheta3) shownValue(j int) Angles {
	if !n.started {
		return n.start(j)
	}
	return n.shown[j-1]
}

func (n *NodePhiTheta3) exchange(ctx context.Context) (in [nodeCount]pathMsg, ok bool) {
	a, b := n.Partners[0], n.Partners[1]
	me := n.Me - 1
	outA := pathMsg{K: n.Card.K[a-1], A: n.sendValue(a)}
	outB := pathMsg{K: n.Card.K[b-1], A: n.sendValue(b)}

	sent, got := [2]bool{}, [2]bool{}
	for !(sent[0] && sent[1] && got[0] && got[1]) {
		var toA, toB chan<- pathMsg
		if !sent[0] {
			toA = n.mesh.link[me][a-1]
		}
		if !sent[1] {
			toB = n.mesh.link[me][b-1]
		}
		var fromA, fromB <-chan pathMsg
		if !got[0] {
			fromA = n.mesh.link[a-1][me]
		}
		if !got[1] {
			fromB = n.mesh.link[b-1][me]
		}

		select {
		case toA <- outA:
			sent[0] = true
		case toB <- outB:
			sent[1] = true
		case v := <-fromA:
			in[a-1], got[0] = v, true
		case v := <-fromB:
			in[b-1], got[1] = v, true
		case e := <-n.EditIn:
			n.applyEdit(e)
		case <-n.StepIn:
			n.steps++
		case <-n.ResetIn:
			n.reset = true
		case <-n.StartIn:
			n.applyStart()
		case <-ctx.Done():
			return in, false
		}
	}
	return in, true
}

func (n *NodePhiTheta3) Update(ctx context.Context) {
	clk := n.Clock.Copy()
	clk.SpeedFrom(n.SpeedCh)
	clk.WakeOn(n.Wake)
	n.geom.Clocks().Use(clk)
	n.postTicks()
	n.postArrows()
	n.placePartners()

	for {
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
		n.drainEdits()
		if n.reset {
			n.applyReset()
		}

		if clk.Speed() > 0 || n.steps > 0 {
			if n.steps > 0 {
				n.steps--
			}
			in, ok := n.exchange(ctx)
			if !ok {
				return
			}
			n.round(in)
		}
	}
}

func (a BuildArgs) Geom() *NodeCat.NodeGeometry {
	if a.Deps == nil {
		return nil
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	return ng
}

var Builder = BuilderFor("NodePhiTheta3",
	func(a BuildArgs) (any, error) {
		me, err := strconv.Atoi(a.Name)
		if err != nil || me < 1 || me > nodeCount {
			return nil, fmt.Errorf("NodePhiTheta3: node %q is not one of nodes 1..%d — the card names its nodes by number, so the tab's node ids must be 1, 2 and 3", a.Name, nodeCount)
		}

		n := &NodePhiTheta3{}
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		inbox := a.EditInbox()
		n.EditIn, n.StepIn, n.ResetIn, n.StartIn, n.Wake = inbox.Edits, inbox.Steps, inbox.Resets, inbox.Starts, inbox.Wake
		n.geom = a.Geom()
		n.mesh = a.Mesh()

		n.Me = me
		n.Partners = CardPanel.Partners(me)
		n.wireTips()
		n.Card = CardPanel.CardFromState(me, a.State())
		n.S = a.S()
		n.M = a.M()
		n.Card.ClampTicks(n.S)
		n.nodeR = NodeCat.NodeRadius(n.geom.Kind())

		n.breadcrumb("built", fmt.Sprintf("node=%d partners=%v s=%d card=%+v", me, n.Partners, n.S, n.Card))

		return n, nil
	})
