package Node

import "github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"

type Lead struct {
	TargetID string
	Vec      polarindex.Offset
}

type Leads struct {
	leads  []Lead
	from   polarindex.Index
	sent   bool
	dragTo func(id string) (Deposit, bool)
}

func (n *Messaging) WireLeads(dragTo func(id string) (Deposit, bool)) { n.leads.dragTo = dragTo }

func (n *Messaging) DragDeposit() Deposit { return n.dragIn.deposit }

func (k *KindPosts) PostLeads(leads []Lead) {
	k.post(func(p *KindPost) { p.Leads = &leads })
}

func (g *NodeGeometry) applyLeads(leads []Lead) {
	g.msg.leads.leads = leads
	g.sendLeads(g.ComposedIndex())
}

func (g *NodeGeometry) leadsOnMove(at polarindex.Index) {
	if g.msg.leads.sent && at == g.msg.leads.from {
		return
	}
	g.sendLeads(at)
}

func (g *NodeGeometry) sendLeads(at polarindex.Index) {
	g.msg.leads.from, g.msg.leads.sent = at, true
	if g.msg.leads.dragTo == nil {
		return
	}
	center := NodeWorldPos(g.geom)
	for _, l := range g.msg.leads.leads {
		deposit, ok := g.msg.leads.dragTo(l.TargetID)
		if !ok {
			continue
		}
		idx := TipIndex(center, g.SceneCenter(), l.Vec, g.Constants())
		deposit(Msg{NodeID: l.TargetID, Body: Drag{Target: &idx}})
	}
}
