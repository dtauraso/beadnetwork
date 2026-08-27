package NodePhiTheta

import (
	"context"
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
)

type NodePhiTheta struct {
	Self *Self

	Clock clock.Clock

	SpeedCh <-chan float64

	VectorOut chan<- TiltPanel.TiltVectorMsg
	VectorIn  <-chan TiltPanel.TiltVectorMsg

	Center Turn
	Top    Turn

	Arrival Turn
	Got     bool

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

func (n *NodePhiTheta) step(arrival Turn) {
	moved := false

	if phi := n.Rings.Phi.Next(n.Center.Phi, n.Top.Phi, arrival.Phi); phi != n.Center.Phi {
		n.Center.Phi = phi
		moved = true
	}

	if theta := n.Rings.Theta.Next(n.Center.Theta, n.Top.Theta, arrival.Theta); theta != n.Center.Theta {
		n.Center.Theta = theta
		moved = true
	}

	if moved {
		n.Self.SetCenter(n.Center)
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
				n.Arrival, n.Got = vectorOf(arrival), true
			}
			if n.Got {
				n.step(n.Arrival)
			}
			n.send()
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

		n.Center, n.Rings = a.CenterSeed()
		n.Top = a.TopSeed(n.Center)

		n.Self.Breadcrumb("built", fmt.Sprintf("name=%s seed phi=%d theta=%d r=%d  rings phi=%d theta=%d  in=%v out=%v",
			a.Name, n.Center.Phi, n.Center.Theta, n.Center.R,
			n.Rings.Phi.Whole, n.Rings.Theta.Whole, n.VectorIn != nil, n.VectorOut != nil))

		return n, nil
	})
