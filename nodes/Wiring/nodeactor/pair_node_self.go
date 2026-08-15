package nodeactor

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

type PairNodeSelf struct {
	geom *NodeGeometry

	speedCh <-chan float64
}

func NewPairNodeSelf(geom *NodeGeometry, speedCh <-chan float64) *PairNodeSelf {
	return &PairNodeSelf{geom: geom, speedCh: speedCh}
}

func (p *PairNodeSelf) StartRule(ctx context.Context, clk clock.Clock) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.StartRuleNode(ctx)
	clk.WakeOn(p.geom.RuleWake())
}

func (p *PairNodeSelf) Breadcrumb(label, value string) {
	if p == nil || p.geom == nil || p.geom.tr == nil {
		return
	}

	p.geom.tr.Breadcrumb(label, p.geom.id, "", value)

	id, ok := T.BreadcrumbLabelID(label)
	if !ok {
		return
	}
	p.geom.postSelfEvents([]rowevent.RowEvent{{
		Kind: T.KindBreadcrumb, Label: id, Debug: 1,
		NodeRow: p.geom.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (p *PairNodeSelf) EmitGeometryOnce() {
	if p == nil || p.geom == nil {
		return
	}

	p.geom.clocks.CopyClockSrc()
	if p.geom.tr != nil {
		p.geom.postSelfEvents([]rowevent.RowEvent{{
			Kind: T.KindNodeGeometry, NodeRow: p.geom.NodeRow(),
			PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
		}})
	}
}

func (p *PairNodeSelf) Step(ctx context.Context, tick int64) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.clocks.ApplySpeed(p.speedCh)

	for {
		progressed, cancelled := g.msg.DrainPending(ctx, g.take)
		if cancelled {
			return
		}
		if !progressed {
			break
		}
	}
	g.msg.FlushPending()

	g.drainRuleMesh()

	g.deriveOutEdgeGeometry(tick)

	g.anim.driveOutWires(ctx, tick)

	g.writeStreamFrame(g.drainSelfEvents())
	g.writeOutEdgeFrames(tick)
	g.writeInteriorFrames()
}

func (p *PairNodeSelf) SetTiltIndex(theta, normalTheta, bottomTheta int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.SetTiltIndex(theta, normalTheta, bottomTheta)
	g.persistTiltVectorAngle()
	if g.tr != nil {
		g.emitGeometry()
	}
}

func (p *PairNodeSelf) SetRoundsToParallel(rounds, msgs int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.readout.SetRoundsToParallel(rounds, msgs)
	if g.tr != nil {
		g.emitGeometry()
	}
}

func (p *PairNodeSelf) SetReceivedVector(theta int32, set bool) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.SetReceivedVector(theta, set)
	if g.tr != nil {
		g.emitGeometry()
	}
}

func (p *PairNodeSelf) SetLatticePoints(points int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.SetLatticePoints(points)
	if g.tr != nil {
		g.emitGeometry()
	}
}

func (p *PairNodeSelf) ClearOutBeads() {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.anim.ClearOutWires()
}
