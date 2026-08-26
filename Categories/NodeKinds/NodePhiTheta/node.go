package NodePhiTheta

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
)

type NodePhiTheta struct {
	Fire         func()
	EmitGeometry func()

	Self *Self

	Clock clock.Clock

	SpeedCh <-chan float64

	In  *beadanimation.Receiver
	Out *beadanimation.Sender

	VectorOut chan<- TiltPanel.TiltVectorMsg
	VectorIn  <-chan TiltPanel.TiltVectorMsg

	Center Turn
	Top    Turn
	Turn   Turn
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
	next := CenterNext(n.Center, n.Top, arrival, n.Turn)
	if next != n.Center {
		n.Center = next
		n.Self.SetCenter(n.Center)
	}

	n.send()
}

func (n *NodePhiTheta) Update(ctx context.Context) {
	tryEmit(n.EmitGeometry)

	clk := n.Clock.Copy()
	clk.SpeedFrom(n.SpeedCh)
	n.Self.StartRule(ctx, clk)

	n.send()

	for {
		if ctx.Err() != nil {
			return
		}

		if arrival, ok := TiltPanel.PollRecvVector(n.VectorIn); ok {
			if n.Fire != nil {
				n.Fire()
			}
			n.step(vectorOf(arrival))
		}

		n.Self.Step(ctx, clk.Tick())
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

var Builder = BuilderFor("NodePhiTheta",
	func(a BuildArgs) (any, error) {
		n := &NodePhiTheta{}
		n.Fire = a.Fire()
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.In = a.In("In")
		n.Out = a.Out("Out")
		n.Self = claimSelfDrive(a)

		n.VectorOut = a.VectorOut()
		n.VectorIn = a.VectorIn()

		n.Center, n.Turn = a.CenterSeed()
		n.Top = n.Center

		return n, nil
	})
