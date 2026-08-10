// build_args_clock.go — BuildArgs methods for a node's own CLOCK origin, tick reader, and
// speed-delivery channel. Split out of build_args.go — see that file's header.

package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// Clock returns the loader's clock ORIGIN, or nil on a test build with no loader. The
// owning goroutine Copy()s it exactly once at its own start — this hands over the origin,
// not a per-goroutine clock.
func (a BuildArgs) Clock() clock.Clock { return a.pb.clock }

// Tick returns a read of the loader clock's current tick, or nil when there is no clock.
func (a BuildArgs) Tick() func() int64 {
	if a.pb.clock == nil {
		return nil
	}
	clk := a.pb.clock
	return func() int64 { return clk.Tick() }
}

// SpeedCh allocates THIS node's buffered-1 speed-delivery channel and registers it with
// the loader's sink, so a speed change reaches this goroutine's own clock copy. Returns
// nil when there is no sink (test builds with no loader). Call it ONCE per node: each
// call allocates and registers another channel, and only the last one a node keeps would
// ever be drained.
func (a BuildArgs) SpeedCh() <-chan float64 {
	if a.pb.speedSinks == nil {
		return nil
	}
	speedCh := make(chan float64, 1)
	*a.pb.speedSinks = append(*a.pb.speedSinks, speedCh)
	return speedCh
}
