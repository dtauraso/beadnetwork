// geom_seeds.go — the load-time seed-geometry owner split out of MoveDispatch (god-object
// decomposition), as a pure move (no logic changes): geomSeeds owns nodeSeeds/edgeSeeds
// and their NodeSeeds/EdgeSeeds/loadTimeCenters accessors. MoveDispatch keeps its public
// NodeSeeds/EdgeSeeds methods as thin delegators so the external API is unchanged.

package Wiring

// geomSeeds owns every node/edge's load-time seed geometry, captured ONCE at
// construction (newMoveDispatch) in spec order — the deterministic directory-sorted
// order LoadTopology read the topology in, NOT map iteration order. Exposed via
// NodeSeeds/EdgeSeeds so main.go can seed the buffer's row tables from the diagram
// itself before any node goroutine starts (CLAUDE.md: rows are a projection of the
// diagram, not a discovery log built by racing goroutines to their first emit).
type geomSeeds struct {
	nodeSeeds []NodeGeomSeed
	edgeSeeds []EdgeGeomSeed
}

// nodeSeedsFn returns every node's load-time seed geometry in SPEC ORDER. Call after
// LoadTopology returns, before launching any node goroutine, and stream each entry via
// tr.NodeGeometry (main.go).
func (gs *geomSeeds) nodeSeedsFn() []NodeGeomSeed { return gs.nodeSeeds }

// edgeSeedsFn returns every edge's load-time seed topology (with real endpoint geometry)
// in SPEC ORDER. Call alongside nodeSeedsFn; stream each entry via tr.Geometry (main.go).
func (gs *geomSeeds) edgeSeedsFn() []EdgeGeomSeed { return gs.edgeSeeds }

// loadTimeCenters returns the node-id → LOAD-TIME world center map, rebuilt from
// gs.nodeSeeds (frozen at construction, in newMoveDispatch, and never mutated
// afterward). Used only by LoadSceneSphere's content-fit fallback, which runs on the
// main goroutine before Start launches any mover goroutine — nodeSeeds is already
// fully populated by then, so this is a safe read.
func (gs *geomSeeds) loadTimeCenters() map[string]vec3 {
	out := make(map[string]vec3, len(gs.nodeSeeds))
	for _, sd := range gs.nodeSeeds {
		out[sd.ID] = vec3{X: sd.CX, Y: sd.CY, Z: sd.CZ}
	}
	return out
}
