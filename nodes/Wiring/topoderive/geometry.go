package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func ComputeNodeGeometry(spec loadspec.TopoSpec, sphere polar.SceneSphere) (map[string]nodegeom.NodeGeom, map[string]spatial.Vec3) {
	nodeGeoms := map[string]nodegeom.NodeGeom{}
	for _, n := range spec.Nodes {
		nodeGeoms[n.ID] = n.ToNodeGeom(sphere.Center)
	}

	centers := map[string]spatial.Vec3{}
	for id, g := range nodeGeoms {
		if g.HasPos {
			centers[id] = nodegeom.NodeWorldPos(g)
		}
	}
	return nodeGeoms, centers
}
