package nodeactor

import (
	"context"

	"github.com/dtauraso/wirefold/src/Clock"
	"github.com/dtauraso/wirefold/src/Node/rowevent"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
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
	p.geom.anim.StartAnimation(ctx)
	clk.WakeOn(p.geom.RuleWake())
}

func (p *PairNodeSelf) Breadcrumb(label, value string) {
	if p == nil || p.geom == nil {
		return
	}

	id, ok := B.BreadcrumbLabelID(label)
	if !ok {
		return
	}
	p.geom.postSelfEvents([]rowevent.RowEvent{{
		Kind: B.KindBreadcrumb, Label: id, Debug: 1,
		NodeRow: p.geom.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (p *PairNodeSelf) EmitGeometryOnce() {
	if p == nil || p.geom == nil {
		return
	}

	p.geom.clocks.CopyClockSrc()
}

func (p *PairNodeSelf) Step(ctx context.Context, tick int64) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.clocks.ApplySpeed(p.speedCh)

	g.beads.ApplyBeadDrag()
}

func (g *NodeGeometry) RunGeometry(ctx context.Context) {
	if g == nil {
		return
	}
	clk := clock.NewRealClock()

	for {
		if ctx.Err() != nil {
			return
		}

		for {
			progressed, cancelled := g.msg.DrainPending(ctx, g.take)
			if cancelled {
				return
			}
			if !progressed {
				break
			}
		}

		g.applyKindPosts()

		g.pollChannelVectors()
		g.drainRuleMesh()

		g.deriveOutEdgeGeometry()

		g.writeStreamFrame(g.drainSelfEvents())
		g.writeOutEdgeFrames(clk.Tick())

		g.writeInteriorFrames()

		if err := clk.SleepPulse(ctx); err != nil {
			return
		}
	}
}

func (p *PairNodeSelf) SetTiltIndex(theta int32) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.kindPosts.PostTiltIndex(theta)
}

func (p *PairNodeSelf) SetRoundsToParallel(rounds, msgs int32) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.kindPosts.PostRoundsToParallel(rounds, msgs)
}

func (p *PairNodeSelf) SetReceivedVector(theta int32, set bool) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.kindPosts.PostReceivedVector(theta, set)
}

func (p *PairNodeSelf) SetLatticePoints(points int32) {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.kindPosts.PostLatticePoints(points)
}

func (g *NodeGeometry) applyKindPosts() {
	p, ok := g.kindPosts.Take()
	if !ok {
		return
	}
	if p.Tilt != nil {
		g.tilt.SetTiltIndex(p.Tilt.Theta)
		g.persistTiltVectorAngle()
	}
	if p.Received != nil {
		g.tilt.SetReceivedVector(p.Received.Theta, p.Received.Set)
	}
	if p.Rounds != nil {
		g.readout.SetRoundsToParallel(p.Rounds.Rounds, p.Rounds.Msgs)
	}
	if p.Lattice != nil {
		g.tilt.SetLatticePoints(*p.Lattice)
	}
	g.emitGeometry()
}

func (p *PairNodeSelf) ClearOutBeads() {
	if p == nil || p.geom == nil {
		return
	}
	p.geom.anim.ClearBeadRuns()
}
