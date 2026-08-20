package topoderive

import (
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/runtopology/loadspec"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/spatial"
)

func ComputeNodeGeometry(spec loadspec.TopoSpec, sphere polar.SceneSphere) (map[string]nodegeom.NodeGeom, map[string]spatial.Vec3) {
	nodeGeoms := map[string]nodegeom.NodeGeom{}
	for _, n := range spec.Nodes {
		nodeGeoms[n.ID] = n.ToNodeGeom(sphere.Center, spec.Constants)
	}

	centers := map[string]spatial.Vec3{}
	for id, g := range nodeGeoms {
		if g.HasPos {
			centers[id] = nodegeom.NodeWorldPos(g)
		}
	}
	return nodeGeoms, centers
}
