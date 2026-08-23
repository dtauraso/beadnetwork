package NormalSum

import (
	"context"

	clock "github.com/dtauraso/wirefold/Categories/Clock"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
	Speed "github.com/dtauraso/wirefold/Categories/Speed"
)

const noNormal = -1

type Node struct {
	Fire  func()
	Clock clock.Clock

	Self *Self

	Points int32

	NormalA *beadanimation.Receiver
	NormalB *beadanimation.Receiver
	Out     *beadanimation.Sender
	SpeedCh <-chan float64

	LatticeIn <-chan int32

	a, b, total int32
}

func (n *Node) Update(ctx context.Context) {
	c := n.Clock.Copy()
	n.Self.StartRule(ctx, c)
	n.a, n.b, n.total = noNormal, noNormal, noNormal

	n.Self.EmitGeometryOnce()

	for {
		if ctx.Err() != nil {
			return
		}
		Speed.ApplySpeedNonBlocking(c, n.SpeedCh)
		select {
		case pts := <-n.LatticeIn:
			if pts > 0 {
				n.Points = pts
			}
		default:
		}

		changed := false

		for {
			v, ok := n.NormalA.PollRecv()
			if !ok {
				break
			}
			n.a = int32(v)
			changed = true
		}
		for {
			v, ok := n.NormalB.PollRecv()
			if !ok {
				break
			}
			n.b = int32(v)
			changed = true
		}

		if changed {
			n.Fire()
			n.republish()
		}

		n.Self.Step(ctx, c.Tick())
		if err := c.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func (n *Node) republish() {
	if n.a == noNormal || n.b == noNormal {
		return
	}
	total := sumIndex(n.a, n.b, n.Points)
	if total == n.total {
		return
	}
	n.total = total

	n.Self.SetTiltIndex(total)

	n.Out.PlaceDrivenAt(int(total), n.Clock.Tick())
}

func sumIndex(a, b, points int32) int32 {
	return wrapIndex(a+b, points)
}

func wrapIndex(i, points int32) int32 {
	if points <= 0 {
		return 0
	}
	i %= points
	if i < 0 {
		i += points
	}
	return i
}

var Builder = BuilderFor("NormalSum",
	func(a BuildArgs) (portwiring.Node, error) {
		n := &Node{}
		n.Fire = a.Fire()
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.Self = claimSelfDrive(a)
		n.Points = a.LatticePointsSeed()
		n.LatticeIn = a.LatticeIn()
		n.NormalA = a.In("NormalA")
		n.NormalB = a.In("NormalB")
		n.Out = a.Out("Out")
		return n, nil
	})
