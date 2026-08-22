package time

import (
	"context"

	lattice "github.com/dtauraso/wirefold/Node/BeadAnimation/lattice"

	clock "github.com/dtauraso/wirefold/Clock"
	beadanimation "github.com/dtauraso/wirefold/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/NodeKinds/nodeapi"

	"github.com/dtauraso/wirefold/Node/nodeactor"
	"github.com/dtauraso/wirefold/NodeKinds/gatecommon"
	Wiring "github.com/dtauraso/wirefold/NodeKinds/kindapi"
)

type Time struct {
	Fire         func()
	EmitGeometry func()
	EmitHeldBead func(held int)
	Held         int `wire:"data.state"`

	Self *nodeactor.PairNodeSelf

	Clock clock.Clock

	SpeedCh <-chan float64

	In     *beadanimation.Receiver
	ToNext beadanimation.Broadcast
}

func placeHeld(outs beadanimation.Broadcast, held int, items []beadanimation.DriveItem, tick int64) []beadanimation.DriveItem {
	if held == gatecommon.NoValue {
		return items
	}
	return outs.PlaceDrivenAllAt(held, items, tick)
}

func (in *Time) drainInput() {
	for {
		if _, ok := in.In.PollRecv(); !ok {
			return
		}
	}
}

func (in *Time) consumeInput(clk clock.Clock, value int, held int) (newHeld int, windowActive bool, windowEndTick int64) {
	if in.Fire != nil {
		in.Fire()
	}

	heldChanged := value != held
	newHeld = value
	if heldChanged && in.EmitHeldBead != nil {
		in.EmitHeldBead(value)
	}

	placeTick := clk.Tick()
	var items []beadanimation.DriveItem
	prevHeld := in.Held
	items = placeHeld(in.ToNext, prevHeld, items, placeTick)
	in.Held = value

	var maxTicks float64
	anyLive := false
	for i, di := range items {
		if !di.Live() {
			continue
		}
		anyLive = true
		if t := float64(in.ToNext[i].Geom().Steps) * lattice.PulsesPerSlot; t > maxTicks {
			maxTicks = t
		}
	}
	if anyLive {
		windowActive = true
		windowEndTick = placeTick + int64(maxTicks+0.999999)
	}
	return newHeld, windowActive, windowEndTick
}

func (in *Time) Update(ctx context.Context) {
	nodeapi.TryEmit(in.EmitGeometry)
	in.Self.EmitGeometryOnce()

	held := gatecommon.NoValue

	if in.EmitHeldBead != nil {
		in.EmitHeldBead(held)
	}

	clk := in.Clock.Copy()
	in.Self.StartRule(ctx, clk)

	windowActive := false
	var windowEndTick int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		clock.ApplySpeedNonBlocking(clk, in.SpeedCh)
		in.Self.Step(ctx, clk.Tick())
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}

		if windowActive {
			in.drainInput()
		} else if value, ok := in.In.PollRecv(); ok {
			held, windowActive, windowEndTick = in.consumeInput(clk, value, held)
		}

		if windowActive && clk.Tick() >= windowEndTick {
			windowActive = false
		}
	}
}

func init() {

	Wiring.RegisterBuilder("Time",
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &Time{

				Held: a.StateSeed("held", gatecommon.NoValue),
			}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.Self = a.ClaimSelfDrive()
			n.In = a.In("In")
			n.ToNext = a.Broadcast("ToNext")

			return n, nil
		})
}
