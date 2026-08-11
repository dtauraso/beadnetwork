// Package geomseeds holds the load-time seed-geometry owner split out of MoveDispatch
// (god-object decomposition), as a pure move (no logic changes): GeomSeeds owns
// NodeSeeds/EdgeSeeds and their NodeSeedsFn/EdgeSeedsFn/LoadTimeCenters accessors.
// MoveDispatch keeps its public NodeSeeds/EdgeSeeds methods as thin delegators to its
// exported GS field so the external API is unchanged.
package geomseeds

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
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

// BuildNodeSeed computes one node's own NodeGeomSeed from its load-time geometry — pulled
// out of newMoveDispatch's node loop (nodes/Wiring/move_dispatch_construct.go), which
// never touched *MoveDispatch here: every input is a parameter (id, its position in spec
// order, and its own geometry) and the only side effect at the old call site was
// appending the return value to md.GS.NodeSeeds. row falls back to i (position in spec
// order) only for a non-numeric id, which real loadTree-built specs never produce; ROW ID
// = NODE ID - 1 is still the rule for every real id.
func BuildNodeSeed(id string, i int, g nodegeom.NodeGeom, row int) NodeGeomSeed {
	label := g.Label
	if label == "" {
		label = id
	}
	var cx, cy, cz float64
	if g.HasPos {
		c := nodegeom.NodeWorldPos(g)
		cx, cy, cz = c.X, c.Y, c.Z
	}
	return NodeGeomSeed{
		ID: id, Label: label, Kind: g.Kind,
		CX: cx, CY: cy, CZ: cz,
		Radius: nodegeom.NodeRadius(g.Kind), SphereR: nodegeom.EffectiveRadius(g),
		VRX: loadspec.VerticalRingNormalX, VRY: loadspec.VerticalRingNormalY, VRZ: loadspec.VerticalRingNormalZ,
		FRX: loadspec.FlatRingNormalX, FRY: loadspec.FlatRingNormalY, FRZ: loadspec.FlatRingNormalZ,
		Row: row,
	}
}

// BuildEdgeSeed computes one edge's own EdgeGeomSeed from its load-time endpoints —
// pulled out of newMoveDispatch's edge loop, same file. geoms is the full node-geometry
// map (a parameter, not md.mr.nodeGeoms) so this stays a pure function of its inputs; the
// missing-endpoint case is reported as an error rather than a panic since a stale edge
// file left behind after its target node's directory was hand-deleted is malformed
// input, not a code bug (in-edges are not indexed, so nothing else catches this at load).
func BuildEdgeSeed(label string, ep inputcodec.EdgeEndpoints, geoms map[string]nodegeom.NodeGeom) (EdgeGeomSeed, error) {
	srcG, srcOK := geoms[ep.Source]
	if !srcOK {
		return EdgeGeomSeed{}, fmt.Errorf("newMoveDispatch: edge %q references missing source node %q (no geometry loaded for it)", label, ep.Source)
	}
	dstG, dstOK := geoms[ep.Target]
	if !dstOK {
		return EdgeGeomSeed{}, fmt.Errorf("newMoveDispatch: edge %q references missing target node %q (no geometry loaded for it)", label, ep.Target)
	}
	seg := nodegeom.EdgeSegment(srcG, dstG)
	return EdgeGeomSeed{
		Label: label, SrcNode: ep.Source, DstNode: ep.Target,
		SX: seg.Start.X, SY: seg.Start.Y, SZ: seg.Start.Z,
		EX: seg.End.X, EY: seg.End.Y, EZ: seg.End.Z,
	}, nil
}

// MutualPairs reports, for every ordered (source, target) pair in edgeEndpoints, whether
// the REVERSE edge (target -> source) also exists — a load-time fact of the edge set
// alone, computed once here with no touch of any node's own state. newMoveDispatch uses
// this to seed each node's topo.mutualTargets so the two chains of a mutual pair offset
// to opposite sides (nodegeom.ParallelChainOffset) instead of drawing on the same line.
func MutualPairs(edgeEndpoints map[string]inputcodec.EdgeEndpoints) map[string]map[string]bool {
	hasEdge := make(map[string]bool, len(edgeEndpoints))
	for _, ep := range edgeEndpoints {
		hasEdge[ep.Source+"\x00"+ep.Target] = true
	}
	out := map[string]map[string]bool{}
	for _, ep := range edgeEndpoints {
		if !hasEdge[ep.Target+"\x00"+ep.Source] {
			continue
		}
		if out[ep.Source] == nil {
			out[ep.Source] = map[string]bool{}
		}
		out[ep.Source][ep.Target] = true
	}
	return out
}
