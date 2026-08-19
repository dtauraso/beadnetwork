package geomseeds

import (
	"fmt"

	"github.com/dtauraso/wirefold/src/Node/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/spatial"
)

type NodeGeomSeed struct {
	ID, Label, Kind    string
	CX, CY, CZ, Radius float64

	Row int
}

type EdgeGeomSeed struct {
	Label, SrcNode, DstNode string
	SX, SY, SZ, EX, EY, EZ  float64
}

type GeomSeeds struct {
	NodeSeeds []NodeGeomSeed
	EdgeSeeds []EdgeGeomSeed
}

func (gs *GeomSeeds) NodeSeedsFn() []NodeGeomSeed { return gs.NodeSeeds }

func (gs *GeomSeeds) EdgeSeedsFn() []EdgeGeomSeed { return gs.EdgeSeeds }

func (gs *GeomSeeds) LoadTimeCenters() map[string]spatial.Vec3 {
	out := make(map[string]spatial.Vec3, len(gs.NodeSeeds))
	for _, sd := range gs.NodeSeeds {
		out[sd.ID] = spatial.Vec3{X: sd.CX, Y: sd.CY, Z: sd.CZ}
	}
	return out
}

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
		Radius: nodegeom.NodeRadius(g.Kind),
		Row:    row,
	}
}

func BuildEdgeSeed(label string, ep inputcodec.EdgeEndpoints, geoms map[string]nodegeom.NodeGeom) (EdgeGeomSeed, error) {
	srcG, srcOK := geoms[ep.Source]
	if !srcOK {
		return EdgeGeomSeed{}, fmt.Errorf("newMoveDispatch: edge %q references missing source node %q (no geometry loaded for it)", label, ep.Source)
	}
	dstG, dstOK := geoms[ep.Target]
	if !dstOK {
		return EdgeGeomSeed{}, fmt.Errorf("newMoveDispatch: edge %q references missing target node %q (no geometry loaded for it)", label, ep.Target)
	}
	seg := edgegeom.EdgeSegment(srcG, dstG)
	return EdgeGeomSeed{
		Label: label, SrcNode: ep.Source, DstNode: ep.Target,
		SX: seg.Start.X, SY: seg.Start.Y, SZ: seg.Start.Z,
		EX: seg.End.X, EY: seg.End.Y, EZ: seg.End.Z,
	}, nil
}

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
