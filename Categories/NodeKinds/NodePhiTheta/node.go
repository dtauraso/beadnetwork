package NodePhiTheta

import (
	"context"
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type NodePhiTheta struct {
	Self *Self

	Clock clock.Clock

	SpeedCh <-chan float64

	VectorOut chan<- TiltPanel.TiltVectorMsg
	VectorIn  <-chan TiltPanel.TiltVectorMsg

	Center    Turn
	PartnerID string

	Top Turn

	loggedPhi, loggedTheta int
	logged                 bool

	Rings Rings
}

func msgOf(v Turn) TiltPanel.TiltVectorMsg {
	return TiltPanel.TiltVectorMsg{
		PhiIdx: int32(v.Phi), ThetaIdx: int32(v.Theta), RIdx: int32(v.R),
	}
}

func vectorOf(m TiltPanel.TiltVectorMsg) Turn {
	return Turn{Phi: int(m.PhiIdx), Theta: int(m.ThetaIdx), R: int(m.RIdx)}
}

func (n *NodePhiTheta) send() {
	TiltPanel.SendVectorLatestNonBlocking(n.VectorOut, msgOf(n.Center))
}

func (r Ring) logLine(angle string, center, top, arrival, next int) string {
	return fmt.Sprintf("%s c=%d->%d top=%d bot=%d arr=%d own=%d dTop=%d dBot=%d q=%d off=%+d",
		angle, center, next, top, r.Bottom(top), arrival,
		r.DistanceOwn(center, top),
		r.DistanceTop(top, arrival), r.DistanceBottom(top, arrival),
		r.Whole/4, r.Offset(center, top, arrival))
}

func (n *NodePhiTheta) step(arrival Turn) {
	moved := false

	offPhi := n.Rings.Phi.Offset(n.Center.Phi, n.Top.Phi, arrival.Phi)
	offTheta := n.Rings.Theta.Offset(n.Center.Theta, n.Top.Theta, arrival.Theta)

	if !n.logged || offPhi != n.loggedPhi {
		n.Self.Breadcrumb("phi", n.Rings.Phi.logLine("phi", n.Center.Phi, n.Top.Phi, arrival.Phi,
			n.Rings.Phi.Next(n.Center.Phi, n.Top.Phi, arrival.Phi)))
		n.loggedPhi = offPhi
	}
	if !n.logged || offTheta != n.loggedTheta {
		n.Self.Breadcrumb("theta", n.Rings.Theta.logLine("theta", n.Center.Theta, n.Top.Theta, arrival.Theta,
			n.Rings.Theta.Next(n.Center.Theta, n.Top.Theta, arrival.Theta)))
		n.loggedTheta = offTheta
	}
	n.logged = true

	if offPhi != 0 {
		n.Center.Phi += offPhi
		moved = true
	}

	if offTheta != 0 {
		n.Center.Theta += offTheta
		moved = true
	}

	if moved {
		n.Self.StepBy(polarindex.Offset{Phi: offPhi, Theta: offTheta})
	}
}

func (n *NodePhiTheta) Update(ctx context.Context) {
	clk := n.Clock.Copy()
	clk.SpeedFrom(n.SpeedCh)
	n.Self.StartRule(ctx, clk)

	n.send()

	for {
		if ctx.Err() != nil {
			return
		}

		if clk.Speed() > 0 {
			if arrival, ok := TiltPanel.PollRecvVector(n.VectorIn); ok {
				n.step(vectorOf(arrival))
				n.send()
			}
		}

		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

var Builder = BuilderFor("NodePhiTheta",
	func(a BuildArgs) (any, error) {
		n := &NodePhiTheta{}
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.Self = claimSelfDrive(a)

		n.VectorOut = a.VectorOut()
		n.VectorIn = a.VectorIn()

		n.PartnerID, n.Center, n.Rings = a.CenterSeed()
		n.Top = a.TopSeed(n.Center)

		n.Self.Breadcrumb("built", fmt.Sprintf("name=%s seed phi=%d theta=%d r=%d  rings phi=%d theta=%d  in=%v out=%v",
			a.Name, n.Center.Phi, n.Center.Theta, n.Center.R,
			n.Rings.Phi.Whole, n.Rings.Theta.Whole, n.VectorIn != nil, n.VectorOut != nil))

		return n, nil
	})
