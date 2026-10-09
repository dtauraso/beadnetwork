package NodePhiTheta3

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

func (n *NodePhiTheta3) applyEdit(e CardPanel.EditMsg) {
	if e.Field.Vector == CardPanel.VecS {
		n.S = e.Value
		n.clearTicks()
		n.postTicks()
		n.postVectors()
		return
	}
	if e.Field.Vector == CardPanel.VecM {
		n.M = e.Value
		return
	}
	n.Card.Set(e.Field, e.Value)
	if v := e.Field.Vector; v == CardPanel.VecStart || v == CardPanel.VecTick || v == CardPanel.VecK {
		n.postVectors()
	}
	n.persist(e.Field, e.Value)
}

func (n *NodePhiTheta3) persist(f CardPanel.Field, value int) {
	if err := NodeCat.WriteCardState(n.geom.PersistRoot(), n.geom.ID(), f.StateKey(), value); err != nil {
		n.breadcrumb("card-persist", err.Error())
	}
}

func (n *NodePhiTheta3) clearTicks() {
	n.Card.Tick = [nodeCount]Vec{}
	for _, f := range CardPanel.TickFields(n.Me) {
		n.persist(f, 0)
	}
}

func (n *NodePhiTheta3) drainEdits() {
	for {
		select {
		case e := <-n.EditIn:
			n.applyEdit(e)
		case <-n.StepIn:
			n.steps++
		case <-n.ResetIn:
			n.reset = true
		case <-n.StartIn:
			n.applyStart()
		default:
			return
		}
	}
}

func (n *NodePhiTheta3) applyReset() {
	n.reset, n.steps, n.started, n.loggedOnce = false, 0, false, false
	n.arrival = [nodeCount]Angles{}
	n.postArrows()
	n.breadcrumb("card", "reset to round 0")
}

func (n *NodePhiTheta3) applyStart() {
	n.placeChain()
	n.breadcrumb("card", "start: placing partners")
}
