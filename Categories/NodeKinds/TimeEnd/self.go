package timeend

import (
	"context"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
)

type Self struct {
	geom *NodeCat.NodeGeometry

	speedCh <-chan float64
}

func NewSelf(geom *NodeCat.NodeGeometry, speedCh <-chan float64) *Self {
	return &Self{geom: geom, speedCh: speedCh}
}

func (p *Self) StartRule(ctx context.Context, clk clock.Clock) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.StartRuleNode(ctx)
	p.geom.Anim().StartBeadAnimation(ctx)
	clk.WakeOn(p.geom.RuleWake())
}

func (p *Self) Breadcrumb(label, value string) {
	if p == nil || p.geom == nil {
		return
	}

	p.geom.Trace().Post([]NodeCat.RowEvent{{
		Kind: NodeCat.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: p.geom.Stream().NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (p *Self) EmitGeometryOnce() {
	if p == nil || p.geom == nil {
		return
	}

	p.geom.Clocks().CopyClockSrc()
}

func (p *Self) Step(ctx context.Context, tick int64) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.Clocks().ApplySpeed(p.speedCh)

	g.Beads().ApplyBeadDrag()
}

func (p *Self) SetKindRule(trim NodeDrag.Trim, request NodeDrag.Request) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.SetKindRule(trim, request)
}

func (p *Self) SetTiltIndex(theta int32) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.KindPosts().PostTiltIndex(theta)
}

func (p *Self) SetRoundsToParallel(rounds, msgs int32) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.KindPosts().PostRoundsToParallel(rounds, msgs)
}

func (p *Self) SetReceivedVector(theta int32, set bool) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.KindPosts().PostReceivedVector(theta, set)
}

func (p *Self) SetLatticePoints(points int32) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.KindPosts().PostLatticePoints(points)
}

func (p *Self) ClearOutBeads() {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.Anim().ClearBeadLines()
}
