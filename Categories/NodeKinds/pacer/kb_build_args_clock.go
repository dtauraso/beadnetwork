package pacer

import clock "github.com/dtauraso/wirefold/Categories/Clock"

func (a BuildArgs) Clock() clock.Clock { return a.PB.Clock }

func (a BuildArgs) SpeedCh() <-chan float64 {
	if a.PB.SpeedSinks == nil {
		return nil
	}
	speedCh := make(chan float64, 1)
	a.PB.SpeedSinks.Clocks = append(a.PB.SpeedSinks.Clocks, speedCh)
	return speedCh
}
