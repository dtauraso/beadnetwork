// Package geomseeds holds the load-time seed-geometry owner split out of MoveDispatch
// (god-object decomposition), as a pure move (no logic changes): GeomSeeds owns
// NodeSeeds/EdgeSeeds and their NodeSeedsFn/EdgeSeedsFn/LoadTimeCenters accessors.
// MoveDispatch keeps its public NodeSeeds/EdgeSeeds methods as thin delegators to its
// exported GS field so the external API is unchanged.
package geomseeds

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// NodeGeomSeed is one node's load-time seed geometry, exported in spec order and consumed
// by main.go's pre-launch tr.NodeGeometry loop (see the row-seeding comment in main.go).
// No port geometry rides here any more (docs/bead-model/channels-not-ports.md — a port carries none).
type NodeGeomSeed struct {
	ID, Label, Kind              string
	CX, CY, CZ, Radius, SphereR  float64
	VRX, VRY, VRZ, FRX, FRY, FRZ float64
	// Row is this node's buffer ROW INDEX: id-1 (ROW ID = NODE ID - 1 — the row is declared
	// by the id, never derived by sorting/position in this slice). Two nodes are never at
	// the same Row (loadTree rejects a duplicate id); a gap in the id space is simply a row
	// no seed claims, not a shift of later rows.
	Row int
}

// EdgeGeomSeed is one edge's load-time topology AND its real segment endpoints — the same
// nodegeom.EdgeSegment(srcGeom, dstGeom) computation the edge's own live recomputeGeometry
// (edge_mover.go) uses, evaluated here against the load-time geoms so the seed row is never a
// degenerate 0,0,0→0,0,0 segment.
type EdgeGeomSeed struct {
	Label, SrcNode, DstNode string
	SX, SY, SZ, EX, EY, EZ  float64
}

// GeomSeeds owns every node/edge's load-time seed geometry, captured ONCE at
// construction (newMoveDispatch) in spec order — the deterministic directory-sorted
// order LoadTopology read the topology in, NOT map iteration order. Exposed via
// NodeSeedsFn/EdgeSeedsFn so main.go can seed the buffer's row tables from the diagram
// itself before any node goroutine starts (CLAUDE.md: rows are a projection of the
// diagram, not a discovery log built by racing goroutines to their first emit).
type GeomSeeds struct {
	NodeSeeds []NodeGeomSeed
	EdgeSeeds []EdgeGeomSeed
}

// NodeSeedsFn returns every node's load-time seed geometry in SPEC ORDER. Call after
// LoadTopology returns, before launching any node goroutine, and stream each entry via
// tr.NodeGeometry (main.go).
func (gs *GeomSeeds) NodeSeedsFn() []NodeGeomSeed { return gs.NodeSeeds }

// EdgeSeedsFn returns every edge's load-time seed topology (with real endpoint geometry)
// in SPEC ORDER. Call alongside NodeSeedsFn; stream each entry via tr.Geometry (main.go).
func (gs *GeomSeeds) EdgeSeedsFn() []EdgeGeomSeed { return gs.EdgeSeeds }

// LoadTimeCenters returns the node-id → LOAD-TIME world center map, rebuilt from
// gs.NodeSeeds (frozen at construction, in newMoveDispatch, and never mutated
// afterward). Used only by LoadSceneSphere's content-fit fallback, which runs on the
// main goroutine before Start launches any mover goroutine — NodeSeeds is already
// fully populated by then, so this is a safe read.
func (gs *GeomSeeds) LoadTimeCenters() map[string]wire.Vec3 {
	out := make(map[string]wire.Vec3, len(gs.NodeSeeds))
	for _, sd := range gs.NodeSeeds {
		out[sd.ID] = wire.Vec3{X: sd.CX, Y: sd.CY, Z: sd.CZ}
	}
	return out
}
