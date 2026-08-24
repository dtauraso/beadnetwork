package timestart

import (
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

func claimSelfDrive(a BuildArgs) *Self {
	if a.Deps == nil {
		return nil
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	if ng == nil {
		return nil
	}

	var speedCh chan float64
	if a.PB.SpeedSinksOf() != nil {
		speedCh = make(chan float64, 1)
		a.PB.SpeedSinksOf().Clocks = append(a.PB.SpeedSinksOf().Clocks, speedCh)

		sliderToAnimSleepCh := make(chan int64, 1)
		a.PB.SpeedSinksOf().Anim = append(a.PB.SpeedSinksOf().Anim, sliderToAnimSleepCh)
		ng.Anim().SetSleepCh(sliderToAnimSleepCh)
	}

	return NewSelf(ng, speedCh)
}
