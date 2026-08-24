package input

import (
	"context"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
)

type Node struct {
	Fire         func()
	EmitGeometry func()

	Self *Self

	EmitNodeBeads func(working, backup []int)

	EmitRefillSlide func(clk clock.Clock, beads []int)

	Clock clock.Clock

	SpeedCh <-chan float64
	Init    []int `wire:"data.init"`
	Repeat  bool  `wire:"data.repeat"`

	OutCadence *beadanimation.Sender

	ToExcitatory *beadanimation.Sender
	FeedbackIn   *beadanimation.Receiver
}

func (n *Node) clock() clock.Clock {
	if n.Clock == nil {
		return clock.NewRealClock()
	}
	return n.Clock
}

func (n *Node) broadcastPlace(v int, tick int64) bool {
	if n.OutCadence.HasRun() && n.OutCadence.PlaceDrivenAt(v, tick).Failed() {
		return false
	}
	if n.ToExcitatory.HasRun() && n.ToExcitatory.PlaceDrivenAt(v, tick).Failed() {
		return false
	}
	return true
}

func (n *Node) Update(ctx context.Context) {
	tryEmit(n.EmitGeometry)
	n.Self.EmitGeometryOnce()

	if len(n.Init) == 0 {
		n.runStepLoop(ctx, n.clock().Copy(), nil)
		return
	}

	init := append([]int(nil), n.Init...)
	working := append([]int(nil), init...)
	backup := append([]int(nil), init...)

	emitBeads := func() {
		if n.EmitNodeBeads != nil {
			n.EmitNodeBeads(working, backup)
		}
	}
	emitBeads()

	clk := n.clock().Copy()
	clk.SpeedFrom(n.SpeedCh)

	if n.FeedbackIn.HasRun() {
		n.updateFeedbackRing(ctx, &working, &backup, init, emitBeads, clk)
		return
	}

	n.runPeriodicEmit(ctx, &working, &backup, init, emitBeads, clk)
}

func (n *Node) runStepLoop(ctx context.Context, clk clock.Clock, perTick func() bool) {
	clk.SpeedFrom(n.SpeedCh)
	n.Self.StartRule(ctx, clk)
	for {
		if ctx.Err() != nil {
			return
		}
		if perTick != nil && !perTick() {

			perTick = nil
		}
		n.Self.Step(ctx, clk.Tick())
		if err := clk.SleepPulse(ctx); err != nil {
			return
		}
	}
}

var Builder = BuilderFor("Input",
	func(a BuildArgs) (any, error) {
		n := &Node{

			Clock: clock.NewRealClock(),
		}
		n.Fire = a.Fire()
		n.EmitNodeBeads = a.EmitNodeBeads()
		n.EmitRefillSlide = a.EmitRefillSlide()

		if clk := a.Clock(); clk != nil {
			n.Clock = clk
		}
		n.SpeedCh = a.SpeedCh()
		n.Self = claimSelfDrive(a)
		n.Self.SetKindRule(trimOwnDrag, equalOutLengths)

		if data := a.Data; data != nil {
			if data.Init != nil {
				n.Init = append([]int(nil), data.Init...)
			}
			n.Repeat = data.Repeat
		}

		n.OutCadence = a.Out("OutCadence")
		n.ToExcitatory = a.Out("ToExcitatory")
		n.FeedbackIn = a.In("FeedbackIn")

		return n, nil
	})
