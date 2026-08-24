package timeend

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

	if a.PB.SpeedSinksOf() != nil {
		sliderToAnimSleepCh := make(chan int64, 1)
		a.PB.SpeedSinksOf().Anim = append(a.PB.SpeedSinksOf().Anim, sliderToAnimSleepCh)
		ng.Anim().SetSleepCh(sliderToAnimSleepCh)
	}

	return NewSelf(ng)
}
