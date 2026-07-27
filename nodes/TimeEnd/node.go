package timeend

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// noValue is the sentinel meaning "no value seen yet" → empty interior.
// Real values are non-negative indices so noValue (-1) never collides.
// Aliases Wiring.NoValue, the one definition (gatecommon.NoValue aliases the
// same constant).
const noValue = Wiring.NoValue

// TimeEnd is the terminal "TimeEnd" kind (registered as "TimeEnd" below). Its
// functions: Update runs the node goroutine — it receives a value on the single
// input, calls Fire, holds the value in Held, and calls EmitHeldBead to re-emit
// the held bead when the value changes; EmitGeometry publishes its geometry. It
// produces NO output — it is the end of a time chain.
type TimeEnd struct {
	wire.LayoutHolder
	Fire         func()
	EmitGeometry func()
	EmitHeldBead func(held int)
	Held         int `wire:"data.state"`
	// Clock is this node's OWN clock storage, seeded by Wiring.reflectBuild
	// directly from the loader's origin (bare-field injection by exact type
	// wire.Clock — see input.Node.Clock; ports no longer hand out a clock,
	// per-goroutine-clock.md API demolition item 1). Update() Copies it exactly
	// once at its own start.
	Clock wire.Clock
	// SpeedCh delivers a speed change to THIS goroutine's own clk copy
	// (per-goroutine-clock.md "Delivery"), seeded by Wiring.reflectBuild
	// (injectSpeedChans). nil on a test build with no loader.
	SpeedCh <-chan float64
	In      *wire.In
}

func (h *TimeEnd) Update(ctx context.Context) {
	wire.TryEmit(h.EmitGeometry)

	held := noValue
	if h.EmitHeldBead != nil {
		h.EmitHeldBead(held)
	}

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine).
	clk := h.Clock.Copy()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wire.ApplySpeedNonBlocking(clk, h.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}

		if value, ok := h.In.PollRecv(); ok {
			if h.Fire != nil {
				h.Fire()
			}
			if value != held && h.EmitHeldBead != nil {
				h.EmitHeldBead(value)
			}
			held = value
			h.Held = value
		}
	}
}

func init() {
	// Held defaults to the empty sentinel, not the int zero-value (0 is a real
	// held value). See holdnewsendold for the seed rationale.
	wire.Register("TimeEnd", func() any { return &TimeEnd{Held: noValue} })
}
