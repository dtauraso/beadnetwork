package gatecommon

import (
	"context"
	"time"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/clock"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

func DriveHeld(ctx context.Context, out Wiring.DrivenOut, heldCh <-chan int64, transform func(int64) int, clk clock.Clock, speedCh <-chan float64) {
	go func() {
		paced := out.Paced()
		cur := int64(NoValue)
		tick, sleep, c := driveHeldClock(clk, speedCh)

		var lastPlaceTick int64
		if paced {
			lastPlaceTick = tick()
		}
		for {
			if ctx.Err() != nil {
				return
			}

			select {
			case v := <-heldCh:
				cur = v
			default:
			}

			place := !paced
			if paced {
				if k, known := driveHeldPeriod(out); known {
					place = tick()-lastPlaceTick >= k
				}

			}
			if place {

				placeTick := tick()
				di := out.PlaceDrivenAt(transform(cur), placeTick)
				if di.Failed() {
					return
				}

				if !di.BufferFull() && paced {
					lastPlaceTick = placeTick
				}
			}

			if paced {
				if k, known := driveHeldPeriod(out); known {
					clock.ApplySpeedNonBlocking(c, speedCh)
					if err := c.SleepUntilTick(ctx, lastPlaceTick+k); err != nil {
						return
					}
					continue
				}
			}
			if err := sleep(ctx); err != nil {
				return
			}
		}
	}()
}

func driveHeldClock(clk clock.Clock, speedCh <-chan float64) (tick func() int64, sleep func(context.Context) error, c clock.Clock) {
	ticker := time.NewTicker(clock.TickPeriod)
	tickCh := ticker.C
	sleep = func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tickCh:
			return nil
		}
	}
	tick = func() int64 { return 0 }
	if clk == nil {
		return tick, sleep, nil
	}
	c = clk.Copy()
	sleep = func(ctx context.Context) error {
		clock.ApplySpeedNonBlocking(c, speedCh)
		return c.SleepCycle(ctx)
	}
	tick = c.Tick
	return tick, sleep, c
}

func driveHeldPeriod(out Wiring.DrivenOut) (k int64, known bool) {
	steps := out.Steps()
	if steps <= 0 {
		return 0, false
	}
	k = int64(float64(steps)*lattice.DwellTicksPerBead + 0.999999)
	if k < 1 {
		k = 1
	}
	return k, true
}
