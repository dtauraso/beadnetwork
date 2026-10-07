package Node

import "github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"

type Lead struct {
	TargetID string
	Vec      polarindex.Offset
}

func (k *KindPosts) PostLeads(leads []Lead) {
	k.post(func(p *KindPost) { p.Leads = &leads })
}

func (g *NodeGeometry) applyLeads(leads []Lead) {
	center := NodeWorldPos(g.geom)
	for _, l := range leads {
		idx := TipIndex(center, g.SceneCenter(), l.Vec, g.Constants())
		g.msg.SendMove()(l.TargetID, Msg{NodeID: l.TargetID, Body: Drag{Target: &idx}})
	}
}
