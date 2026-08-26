package NodePhiTheta

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
)

func (a BuildArgs) VectorOut() chan<- TiltPanel.TiltVectorMsg {
	return a.PB.VectorOutOf(a.Name)
}

func (a BuildArgs) VectorIn() <-chan TiltPanel.TiltVectorMsg {
	return a.PB.VectorInOf(a.Name)
}

func (a BuildArgs) CenterSeed() (center, whole Turn) {
	if a.Deps == nil {
		return Turn{}, Turn{}
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	if ng == nil {
		return Turn{}, Turn{}
	}
	c := ng.Constants()
	return ng.ComposedIndex(), Turn{Phi: c.MaxIndexPhi, Theta: c.MaxIndexTheta, R: 0}
}
