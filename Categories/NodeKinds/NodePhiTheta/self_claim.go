package NodePhiTheta

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
	return NewSelf(ng)
}
