package input

import (
	"context"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/nodeapi"
	"github.com/dtauraso/wirefold/nodes/wire/inport"
	"github.com/dtauraso/wirefold/nodes/wire/outport"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

type Node struct {
	Fire         func()
	EmitGeometry func()

	Self *nodeactor.PairNodeSelf

	EmitNodeBeads func(working, backup []int)

	EmitRefillSlide func(clk clock.Clock, speedCh <-chan float64, beads []int)

	Clock clock.Clock

	SpeedCh <-chan float64
	Init    []int `wire:"data.init"`
	Repeat  bool  `wire:"data.repeat"`

	OutCadence *outport.Out

	ToExcitatory *outport.Out
	FeedbackIn   *inport.In
}

func (n *Node) clock() clock.Clock {
	if n.Clock == nil {
		return clock.NewRealClock()
	}
	return n.Clock
}

func (n *Node) broadcastPlace(v int, tick int64) bool {
	if n.OutCadence.Wired() && n.OutCadence.PlaceDrivenAt(v, tick).Failed() {
		return false
	}
	if n.ToExcitatory.Wired() && n.ToExcitatory.PlaceDrivenAt(v, tick).Failed() {
		return false
	}
	return true
}

func (n *Node) Update(ctx context.Context) {
	nodeapi.TryEmit(n.EmitGeometry)
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

	if n.FeedbackIn.Wired() {
		n.updateFeedbackRing(ctx, &working, &backup, init, emitBeads, clk)
		return
	}

	n.runPeriodicEmit(ctx, &working, &backup, init, emitBeads, clk)
}

func (n *Node) runStepLoop(ctx context.Context, clk clock.Clock, perTick func() bool) {
	n.Self.StartRule(ctx, clk)
	for {
		if ctx.Err() != nil {
			return
		}
		if perTick != nil && !perTick() {

			perTick = nil
		}
		clock.ApplySpeedNonBlocking(clk, n.SpeedCh)
		n.Self.Step(ctx, clk.Tick())
		if err := clk.SleepPulse(ctx); err != nil {
			return
		}
	}
}

func init() {

	Wiring.RegisterBuilder("Input",
		[]portwiring.PortSpec{
			{Name: "OutCadence", Dir: portwiring.PortOut},
			{Name: "ToExcitatory", Dir: portwiring.PortOut},
			{Name: "FeedbackIn", Dir: portwiring.PortIn},
		},
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
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
			n.Self = a.ClaimSelfDrive()

			if data := a.Data(); data != nil {
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
}
