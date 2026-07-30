package pacer

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// noValue is the sentinel meaning "no value seen yet". Real values are
// non-negative indices so noValue (-1) never collides with a legitimate value.
// Aliases Wiring.NoValue, the one definition (gatecommon.NoValue aliases the
// same constant).
const noValue = Wiring.NoValue

type Node struct {
	wire.LayoutHolder
	Fire         func()
	EmitGeometry func()
	EmitHeldBead func(held int)
	Held         int `wire:"data.state"`
	// Clock is this node's OWN clock storage, assigned by this kind's own builder
	// directly from the loader's origin (bare-field injection by exact type
	// wire.Clock — see input.Node.Clock; ports no longer hand out a clock,
	// per-goroutine-clock.md API demolition item 1). Update() Copies it exactly
	// once at its own start.
	Clock wire.Clock
	// SpeedCh delivers a speed change to THIS goroutine's own clk copy
	// (per-goroutine-clock.md "Delivery"), assigned by this kind's own builder
	// (injectSpeedChans). nil on a test build with no loader.
	SpeedCh <-chan float64
	// In is the sole input: the value this node samples to compute a change-step
	// (rule 4 — with exactly one input, there is nothing to distinguish it from).
	// Was "FromInput", a kind-leak name naming the "input" kind; the sender does
	// not have to be an Input node.
	In          *wire.In
	FeedbackOut *wire.Out
}

func (p *Node) Update(ctx context.Context) {
	wire.TryEmit(p.EmitGeometry)

	held := noValue
	if p.EmitHeldBead != nil {
		p.EmitHeldBead(held)
	}

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine).
	clk := p.Clock.Copy()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wire.ApplySpeedNonBlocking(clk, p.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}

		if value, ok := p.In.PollRecv(); ok {
			if p.Fire != nil {
				p.Fire()
			}

			heldChanged := value != held
			held = value
			if heldChanged && p.EmitHeldBead != nil {
				p.EmitHeldBead(value)
			}

			// Change-step feedback: 1 when the value changed (or first recv),
			// 0 when it repeats. Placed fire-and-forget on FeedbackOut (no
			// consume acknowledgment, per MODEL.md).
			step := 0
			if heldChanged {
				step = 1
			}
			p.Held = value

			p.FeedbackOut.PlaceDrivenAt(step, clk.Tick())
		}
	}
}

func init() {
	// Pacer CONSTRUCTS ITSELF. Every assignment below was previously performed by
	// Wiring.reflectBuild via reflection — see Time for the general note.
	Wiring.RegisterBuilder("Pacer",
		[]Wiring.PortSpec{
			{Name: "In", Dir: Wiring.PortIn},
			{Name: "FeedbackOut", Dir: Wiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &Node{
				// Held defaults to the empty sentinel, not the int zero-value (0
				// is a real held value). See Time for the seed rationale.
				Held: a.StateSeed("held", noValue),
			}
			n.Fire = a.Fire()
			n.EmitHeldBead = a.EmitHeldBead()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.In = a.In("In")
			n.FeedbackOut = a.Out("FeedbackOut")
			// EmitGeometry stays nil deliberately — nodeMover/edgeMover emit the same
			// geometry from their own goroutine start.
			return n, nil
		})
}
