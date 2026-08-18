package kindapi

import "github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"

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
		*a.pb.SpeedSinks = append(*a.pb.SpeedSinks, speedCh)

		sliderToAnimSpeedCh := make(chan float64, 1)
		*a.pb.SpeedSinks = append(*a.pb.SpeedSinks, sliderToAnimSpeedCh)
		ng.SetAnimSpeedCh(sliderToAnimSpeedCh)
	}

	return nodeactor.NewPairNodeSelf(ng, speedCh)
}
