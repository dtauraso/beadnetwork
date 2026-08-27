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

func (a BuildArgs) CenterSeed() (center Turn, rings Rings) {
	if a.Deps == nil {
		return Turn{}, Rings{}
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	if ng == nil {
		return Turn{}, Rings{}
	}
	c := ng.Constants()
	return ng.ComposedIndex(), RingsFor(c.MaxIndexPhi, c.MaxIndexTheta)
}
