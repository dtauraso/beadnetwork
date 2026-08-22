package Scene

import (
	"github.com/dtauraso/wirefold/src/Node/Edge/edgegeom"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
)

type GeomSeeds struct {
	NodeSeeds []nodegeom.Seed
	EdgeSeeds []edgegeom.Seed
}

func (gs *GeomSeeds) NodeSeedsFn() []nodegeom.Seed { return gs.NodeSeeds }

func (gs *GeomSeeds) EdgeSeedsFn() []edgegeom.Seed { return gs.EdgeSeeds }

func (gs *GeomSeeds) LoadTimeCenters() map[string]Vec3 {
	out := make(map[string]Vec3, len(gs.NodeSeeds))
	for _, sd := range gs.NodeSeeds {
		out[sd.ID] = Vec3{X: sd.CX, Y: sd.CY, Z: sd.CZ}
	}
	return out
}
