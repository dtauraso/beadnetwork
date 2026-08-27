package NodePhiTheta

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (a BuildArgs) VectorOut() chan<- TiltPanel.TiltVectorMsg {
	return a.PB.VectorOutOf(a.Name)
}

func (a BuildArgs) VectorIn() <-chan TiltPanel.TiltVectorMsg {
	return a.PB.VectorInOf(a.Name)
}

func (a BuildArgs) CenterSeed() (partnerID string, center Turn, rings Rings) {
	if a.Deps == nil {
		return "", Turn{}, Rings{}
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	if ng == nil {
		return "", Turn{}, Rings{}
	}
	c := ng.Constants()
	rings = RingsFor(c.MaxIndexPhi, c.MaxIndexTheta)

	for _, id := range ng.OutTargets() {
		if vec, ok := ng.VectorFromPartner(id); ok {
			return id, Turn(vec), rings
		}
	}
	return "", Turn{}, rings
}

func (a BuildArgs) TopSeed(center Turn) Turn {
	phi, theta := polar.WorldAxisPole()
	if a.Deps == nil {
		return Turn{R: center.R}
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	if ng == nil {
		return Turn{R: center.R}
	}
	idx := polarindex.MeasureIndex(polar.Polar{R: 1, Phi: phi, Theta: theta}, ng.Constants())
	return Turn{Phi: idx.Phi, Theta: idx.Theta, R: center.R}
}
