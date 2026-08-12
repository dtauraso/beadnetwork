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

func (p *PairNodeSelf) Breadcrumb(label, value string) {
	if p == nil || p.geom == nil || p.geom.tr == nil {
		return
	}

	p.geom.tr.Breadcrumb(label, p.geom.id, "", value)

	id, ok := T.BreadcrumbLabelID(label)
	if !ok {
		return
	}
	p.geom.writeStreamFrame([]rowevent.RowEvent{{
		Kind: T.KindBreadcrumb, Label: id, Debug: 1,
		NodeRow: p.geom.stream.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (p *PairNodeSelf) EmitGeometryOnce() {
	if p == nil || p.geom == nil {
		return
	}
	if p.geom.tr != nil {
		p.geom.emitGeometry()
	}
}

func (p *PairNodeSelf) Step(ctx context.Context, tick int64) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	clock.ApplySpeedNonBlocking(g.clocks.clk, p.speedCh)
	for {
		progressed := false
		select {
		case msg := <-g.msg.extIn:
			g.handle(msg)
			if msg.TestDone != nil {
				close(msg.TestDone)
			}
			progressed = true
		default:
		}
		for _, ch := range g.msg.neighborIn {
			select {
			case msg := <-ch:
				g.handle(msg)
				if msg.TestDone != nil {
					close(msg.TestDone)
				}
				progressed = true
			default:
			}
		}
		if !progressed {
			break
		}
	}
	for _, pw := range g.outs.outWires {
		pw.DriveOneCycle(ctx, tick)
	}
	g.msg.flushPending()
	g.writeStreamFrame(nil)
}

func (p *PairNodeSelf) SetTiltIndex(theta, normalTheta, bottomTheta int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.topTiltVectorThetaIdx = theta
	g.tilt.normalThetaIdx = normalTheta
	g.tilt.bottomThetaIdx = bottomTheta
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
	g.readout.roundsToParallel = rounds
	g.readout.msgsToParallel = msgs
	if g.tr != nil {
		g.emitGeometry()
	}
}

func (p *PairNodeSelf) SetReceivedVector(theta int32, set bool) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.receivedVectorThetaIdx = theta
	g.tilt.receivedVectorSet = set
	if g.tr != nil {
		g.emitGeometry()
	}
}

func (p *PairNodeSelf) SetLatticePoints(points int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.latticePoints = points
	if g.tr != nil {
		g.emitGeometry()
	}
}

func (p *PairNodeSelf) ClearOutBeads() {
	if p == nil || p.geom == nil {
		return
	}
	for _, pw := range p.geom.outs.outWires {
		pw.ClearInFlight()
	}
}
