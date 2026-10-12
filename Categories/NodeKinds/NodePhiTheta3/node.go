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

type NodePhiTheta3 struct {
	geom *NodeCat.NodeGeometry

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

	Card   CardPanel.Card
	S      int
	M      int
	CW     int
	CCW    int
	cur    Angles

	nodeR float64
}

func (n *NodePhiTheta3) breadcrumb(label, value string) {
	n.geom.Trace().Post([]NodeCat.RowEvent{{
		Kind: NodeCat.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: n.geom.Stream().NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (n *NodePhiTheta3) restart() {
	n.cur = ticksOf(n.Card.Start, n.S)
	n.postVectors()
}

func (n *NodePhiTheta3) move() {
	next := moveToward(n.cur, ticksOf(n.Card.End, n.S), n.S, n.M, n.CW, n.CCW)
	if next == n.cur {
		return
	}
	n.cur = next
	n.postVectors()
	n.breadcrumb("move", fmt.Sprintf("start=%+v end=%+v", n.cur, ticksOf(n.Card.End, n.S)))
}

func (n *NodePhiTheta3) Update(ctx context.Context) {
	clk := n.Clock.Copy()
	clk.SpeedFrom(n.SpeedCh)
	clk.WakeOn(n.Wake)
	n.geom.Clocks().Use(clk)
	n.postTicks()
	n.restart()

	for {
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
		n.drainEdits()
		if n.reset {
			n.reset = false
			n.restart()
		}
		if clk.Speed() > 0 || n.steps > 0 {
			if n.steps > 0 {
				n.steps--
			}
			n.move()
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

		n.Me = me
		n.Partners = CardPanel.Partners(me)
		n.Card = CardPanel.CardFromState(me, a.State())
		n.S, n.M, n.CW, n.CCW = a.S(), a.M(), a.CW(), a.CCW()
		n.nodeR = NodeCat.NodeRadius(n.geom.Kind())

		n.breadcrumb("built", fmt.Sprintf("node=%d partners=%v s=%d m=%d cw=%d ccw=%d card=%+v", me, n.Partners, n.S, n.M, n.CW, n.CCW, n.Card))

		return n, nil
	})
