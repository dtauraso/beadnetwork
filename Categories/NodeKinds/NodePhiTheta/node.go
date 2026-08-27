package NodePhiTheta

import (
	"context"

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

	n.send()
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

		n.Center, n.Rings = a.CenterSeed()
		n.Top = n.Center

		return n, nil
	})
