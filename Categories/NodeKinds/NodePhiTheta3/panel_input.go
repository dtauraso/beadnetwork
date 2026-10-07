package NodePhiTheta3

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

func (n *NodePhiTheta3) applyEdit(e CardPanel.EditMsg) {
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
		case <-n.ResetIn:
			n.reset = true
		default:
			return
		}
	}
}

func (n *NodePhiTheta3) applyReset() {
	n.reset, n.steps, n.started, n.loggedOnce = false, 0, false, false
	n.arrival = [nodeCount]Vec{}
	n.breadcrumb("card", "reset to round 0")
}
