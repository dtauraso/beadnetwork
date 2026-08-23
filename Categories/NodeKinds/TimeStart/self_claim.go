package timestart

import (
	NodeCat "github.com/dtauraso/wirefold/Categories/Node"
	Wiring "github.com/dtauraso/wirefold/Categories/NodeKinds/kindapi"
)

func claimSelfDrive(a Wiring.BuildArgs) *Self {
	if a.Deps.ClaimSelfDriveGeom == nil {
		return nil
	}
	ng, _ := a.Deps.ClaimSelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	if ng == nil {
		return nil
	}

	var speedCh chan float64
	if a.PB.SpeedSinks != nil {
		speedCh = make(chan float64, 1)
		a.PB.SpeedSinks.Clocks = append(a.PB.SpeedSinks.Clocks, speedCh)

		sliderToAnimSleepCh := make(chan int64, 1)
		a.PB.SpeedSinks.Anim = append(a.PB.SpeedSinks.Anim, sliderToAnimSleepCh)
		ng.Anim().SetSleepCh(sliderToAnimSleepCh)
	}

	return NewSelf(ng, speedCh)
}
