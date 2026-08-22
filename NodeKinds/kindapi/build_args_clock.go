package kindapi

import clock "github.com/dtauraso/wirefold/Clock"

func (a BuildArgs) Clock() clock.Clock { return a.pb.Clock }

func (a BuildArgs) SpeedCh() <-chan float64 {
	if a.pb.SpeedSinks == nil {
		return nil
	}
	speedCh := make(chan float64, 1)
	a.pb.SpeedSinks.Clocks = append(a.pb.SpeedSinks.Clocks, speedCh)
	return speedCh
}
