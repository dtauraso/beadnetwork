package pulseright

import clock "github.com/dtauraso/wirefold/Categories/Clock"

func (a BuildArgs) Clock() clock.Clock { return a.PB.ClockOf() }

func (a BuildArgs) SpeedCh() <-chan float64 {
	if a.PB.SpeedSinksOf() == nil {
		return nil
	}
	speedCh := make(chan float64, 1)
	a.PB.SpeedSinksOf().Clocks = append(a.PB.SpeedSinksOf().Clocks, speedCh)
	return speedCh
}
