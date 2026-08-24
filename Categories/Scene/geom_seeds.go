package Scene

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgegeom"
	"github.com/dtauraso/beadnetwork/Categories/Node"
)

type GeomSeeds struct {
	NodeSeeds []Node.Seed
	EdgeSeeds []edgegeom.Seed
}

func (gs *GeomSeeds) NodeSeedsFn() []Node.Seed { return gs.NodeSeeds }

func (gs *GeomSeeds) EdgeSeedsFn() []edgegeom.Seed { return gs.EdgeSeeds }

func (gs *GeomSeeds) LoadTimeCenters() map[string]Vec3 {
	out := make(map[string]Vec3, len(gs.NodeSeeds))
	for _, sd := range gs.NodeSeeds {
		out[sd.ID] = Vec3{X: sd.CX, Y: sd.CY, Z: sd.CZ}
	}
	return out
}
