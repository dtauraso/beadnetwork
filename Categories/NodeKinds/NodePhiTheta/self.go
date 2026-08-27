package NodePhiTheta

import (
	"context"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Self struct {
	geom *NodeCat.NodeGeometry
}

func NewSelf(geom *NodeCat.NodeGeometry) *Self {
	return &Self{geom: geom}
}

func (p *Self) StartRule(ctx context.Context, clk clock.Clock) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.Clocks().Use(clk)
}

func (p *Self) Breadcrumb(label, value string) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.Trace().Post([]NodeCat.RowEvent{{
		Kind: NodeCat.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: p.geom.Stream().NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (p *Self) SetCenter(center polarindex.Index) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.KindPosts().PostCenter(center)
}
