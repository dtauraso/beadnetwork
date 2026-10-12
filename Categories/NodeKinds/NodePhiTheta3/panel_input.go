package NodePhiTheta3

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

func (n *NodePhiTheta3) applyEdit(e CardPanel.EditMsg) {
	switch e.Field.Vector {
	case CardPanel.VecS:
		n.S = e.Value
		n.postTicks()
		n.restart()
		n.place(true)
		return
	case CardPanel.VecM:
		n.M = e.Value
		return
	case CardPanel.VecCW:
		n.CW = e.Value
		return
	case CardPanel.VecCCW:
		n.CCW = e.Value
		return
	}
	n.Card.Set(e.Field, e.Value)
	n.persist(e.Field, e.Value)
	switch e.Field.Vector {
	case CardPanel.VecStart:
		n.restart()
	case CardPanel.VecL:
		n.tell()
		n.place(false)
	}
}

func (n *NodePhiTheta3) persist(f CardPanel.Field, value int) {
	if err := NodeCat.WriteCardState(n.geom.PersistRoot(), n.geom.ID(), f.StateKey(), value); err != nil {
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
		case <-n.StartIn:
			n.place(true)
		default:
			return
		}
	}
}
