package NodePhiTheta3

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

const nodeCount = CardPanel.NodeCount

var link [nodeCount][nodeCount]chan TiltPanel.TiltVectorMsg

func init() {
	for from := 0; from < nodeCount; from++ {
		for to := 0; to < nodeCount; to++ {
			if from != to {
				link[from][to] = make(chan TiltPanel.TiltVectorMsg)
			}
		}
	}
}

type NodePhiTheta3 struct {
	geom *NodeCat.NodeGeometry

	Clock   clock.Clock
	SpeedCh <-chan float64
	EditIn  <-chan CardPanel.EditMsg
	StepIn  <-chan struct{}
	Wake    <-chan struct{}

	steps int

	Me       int
	Partners [nodeCount - 1]int

	Card CardPanel.Card
	S    int
	M    int

	stepsPerR float64

	arrival [nodeCount]Vec
	started bool

	logged     [nodeCount]Vec
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

func (n *NodePhiTheta3) applyEdit(e CardPanel.EditMsg) {
	if e.Field.Vector == CardPanel.VecS {
		n.S = e.Value
		n.postTicks()
		n.postStarts()
		return
	}
	if e.Field.Vector == CardPanel.VecM {
		n.M = e.Value
		return
	}
	n.Card.Set(e.Field, e.Value)
	if e.Field.Vector == CardPanel.VecStart || e.Field.Vector == CardPanel.VecK {
		n.postStarts()
	}
	if err := NodeCat.WriteCardState(n.geom.PersistRoot(), n.geom.ID(), e.Field.StateKey(), e.Value); err != nil {
		n.breadcrumb("card-persist", err.Error())
	}
}

func (n *NodePhiTheta3) drainEdits() {
	for {
		select {
		case e := <-n.EditIn:
			n.applyEdit(e)
		case <-n.StepIn:
			n.steps++
		default:
			return
		}
	}
}

func (n *NodePhiTheta3) sendValue(j int) Vec {
	if !n.started {
		return n.Card.Start[j-1]
	}
	return scale(n.Card.K[j-1], n.arrival[j-1])
}

func (n *NodePhiTheta3) exchange(ctx context.Context) (in [nodeCount]Vec, ok bool) {
	a, b := n.Partners[0], n.Partners[1]
	me := n.Me - 1
	outA, outB := msgOf(n.sendValue(a)), msgOf(n.sendValue(b))

	sent, got := [2]bool{}, [2]bool{}
	for !(sent[0] && sent[1] && got[0] && got[1]) {
		var toA, toB chan<- TiltPanel.TiltVectorMsg
		if !sent[0] {
			toA = link[me][a-1]
		}
		if !sent[1] {
			toB = link[me][b-1]
		}
		var fromA, fromB <-chan TiltPanel.TiltVectorMsg
		if !got[0] {
			fromA = link[a-1][me]
		}
		if !got[1] {
			fromB = link[b-1][me]
		}

		select {
		case toA <- outA:
			sent[0] = true
		case toB <- outB:
			sent[1] = true
		case v := <-fromA:
			in[a-1], got[0] = vecOf(v), true
		case v := <-fromB:
			in[b-1], got[1] = vecOf(v), true
		case e := <-n.EditIn:
			n.applyEdit(e)
		case <-n.StepIn:
			n.steps++
		case <-ctx.Done():
			return in, false
		}
	}
	return in, true
}

func (n *NodePhiTheta3) round(in [nodeCount]Vec) {
	a, b := n.Partners[0], n.Partners[1]
	for _, j := range n.Partners {
		if n.Card.L[j-1] == 0 {
			in[j-1] = Vec{}
		}
	}

	chosen := pickOne(n.Card.K[a-1], n.Card.K[b-1], in[a-1], in[b-1])
	for _, j := range n.Partners {
		n.arrival[j-1] = step(chosen, n.Card.PoleOffset[j-1], n.M, n.S)
	}
	n.started = true

	if !n.loggedOnce || n.arrival != n.logged {
		n.breadcrumb("card", fmt.Sprintf("in %d=%v %d=%v chosen=%v arrival %d=%v %d=%v",
			a, in[a-1], b, in[b-1], chosen, a, n.arrival[a-1], b, n.arrival[b-1]))
		n.logged, n.loggedOnce = n.arrival, true
	}
}

func (n *NodePhiTheta3) Update(ctx context.Context) {
	clk := n.Clock.Copy()
	clk.SpeedFrom(n.SpeedCh)
	clk.WakeOn(n.Wake)
	n.geom.Clocks().Use(clk)
	n.postTicks()
	n.postStarts()

	for {
		if ctx.Err() != nil {
			return
		}
		n.drainEdits()

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

		if err := clk.SleepCycle(ctx); err != nil {
			return
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
		n.EditIn, n.StepIn, n.Wake = inbox.Edits, inbox.Steps, inbox.Wake
		n.geom = a.Geom()

		n.Me = me
		n.Partners = CardPanel.Partners(me)
		n.wirePlacement()
		n.Card = CardPanel.CardFromState(me, a.State())
		n.S = a.S()
		n.M = a.M()
		n.stepsPerR = NodeCat.NodeRadius(n.geom.Kind()) / n.geom.Constants().ConstantR

		n.breadcrumb("built", fmt.Sprintf("node=%d partners=%v s=%d card=%+v", me, n.Partners, n.S, n.Card))

		return n, nil
	})
