package kindapi

import "github.com/dtauraso/wirefold/src/Node/nodeactor"

func (a BuildArgs) ClaimSelfDrive() *nodeactor.PairNodeSelf {
	if a.deps.ClaimSelfDriveGeom == nil {
		return nil
	}
	ng := a.deps.ClaimSelfDriveGeom(a.name)
	if ng == nil {
		return nil
	}

	var speedCh chan float64
	if a.pb.SpeedSinks != nil {
		speedCh = make(chan float64, 1)
		a.pb.SpeedSinks.Clocks = append(a.pb.SpeedSinks.Clocks, speedCh)

		sliderToAnimSleepCh := make(chan int64, 1)
		a.pb.SpeedSinks.Anim = append(a.pb.SpeedSinks.Anim, sliderToAnimSleepCh)
		ng.SetAnimSleepCh(sliderToAnimSleepCh)
	}

	return nodeactor.NewPairNodeSelf(ng, speedCh)
}
