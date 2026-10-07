package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Lead struct {
	TargetID string
	Vec      polarindex.Offset
}

func (k *KindPosts) PostLeads(leads []Lead) {
	k.post(func(p *KindPost) { p.Leads = &leads })
}

func (g *NodeGeometry) applyLeads(leads []Lead) {
	sc := g.Constants()
	center := NodeWorldPos(g.geom)
	for _, l := range leads {
		tip := center.Add(Vec3(polar.Polar2cart(polarindex.OffsetToPolar(l.Vec, sc))))
		idx := polarindex.MeasureIndex(polar.Cart2polar(polar.Vec3(tip.Sub(g.SceneCenter()))), sc)
		g.msg.SendMove()(l.TargetID, Msg{NodeID: l.TargetID, Body: Drag{Target: &idx}})
	}
}
