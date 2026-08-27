package Node

import (
	"context"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

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

		g.DeriveOutEdgeGeometry()

		g.writeStreamFrame(g.drainSelfEvents())
		g.writeOutEdgeFrames(clk.Tick())

		g.writeInteriorFrames()

		if err := clk.SleepPulse(ctx); err != nil {
			return
		}
	}
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
	if p.Center != nil {

		g.msg.ApplyDerived(g.id, *p.Center)
	}
	if p.FromPartner != nil {

		if partner, ok := g.PartnerIndex(p.FromPartner.PartnerID); ok {
			g.msg.ApplyDerived(g.id, polarindex.Compose(partner, p.FromPartner.Vec, g.Constants()))
		}
	}
	g.emitGeometry()
}
