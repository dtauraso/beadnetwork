package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func ComputeNodeGeometry(spec loadspec.TopoSpec, sphere geom.SceneSphere) (map[string]nodegeom.NodeGeom, map[string]wire.Vec3) {
	nodeGeoms := map[string]nodegeom.NodeGeom{}
	for _, n := range spec.Nodes {
		nodeGeoms[n.ID] = n.ToNodeGeom(sphere.Center)
	}

	centers := map[string]wire.Vec3{}
	for id, g := range nodeGeoms {
		if g.HasPos {
			centers[id] = nodegeom.NodeWorldPos(g)
		}
	}
	return nodeGeoms, centers
}
