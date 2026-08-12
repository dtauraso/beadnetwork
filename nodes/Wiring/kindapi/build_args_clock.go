package kindapi

import (
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

func (a BuildArgs) Clock() clock.Clock { return a.pb.Clock }

func (a BuildArgs) Tick() func() int64 {
	if a.pb.Clock == nil {
		return nil
	}
	clk := a.pb.Clock
	return func() int64 { return clk.Tick() }
}

func (a BuildArgs) SpeedCh() <-chan float64 {
	if a.pb.SpeedSinks == nil {
		return nil
	}
	speedCh := make(chan float64, 1)
	*a.pb.SpeedSinks = append(*a.pb.SpeedSinks, speedCh)
	return speedCh
}
