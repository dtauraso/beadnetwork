// geometry.go — the node-geometry derive phase: the id→geometry map used for
// arc-length computation at wire construction, plus the shared world-center map
// reused by the reach-radius pass, the aimed-port registry, and the edge-geometry
// centerOf closure.
//
// Lifted out of nodes/Wiring/build_geometry.go (movedispatch-decomposition.md item 8's
// class: a pure derive phase that touches no md./buildCtx/actor type) — same body,
// same package boundary shape as scenepersist.
package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// ComputeNodeGeometry builds the id→geometry map used for arc-length computation
// at wire construction (nodegeom.NodeGeom carries kind/dims/port side+slot so the Go arc
// length mirrors buildPortCurve exactly), plus the shared world-center map built
// ONCE from that geometry and reused by the reach-radius pass, the aimed-port
// registry, and the edge-geometry centerOf closure. Each node's world center is
// loaded directly from its spec (meta.json x/y/z, injected as nodegeom.NodeGeom.Center in
// toNodeGeom); nothing later mutates a node's Center, so this snapshot stays
// authoritative for the whole build (the reach-radius pass writes ReachR, which
// does not affect centers).
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
