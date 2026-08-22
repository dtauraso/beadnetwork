package edgegeom

import (
	"fmt"

	"github.com/dtauraso/wirefold/Node/nodegeom"
)

type Seed struct {
	Label, SrcNode, DstNode string
	SX, SY, SZ, EX, EY, EZ  float64
}

func NewSeed(label, srcID, dstID string, geoms map[string]nodegeom.NodeGeom) (Seed, error) {
	srcG, ok := geoms[srcID]
	if !ok {
		return Seed{}, fmt.Errorf("edgegeom.NewSeed: edge %q references missing source node %q (no geometry loaded for it)", label, srcID)
	}
	dstG, ok := geoms[dstID]
	if !ok {
		return Seed{}, fmt.Errorf("edgegeom.NewSeed: edge %q references missing target node %q (no geometry loaded for it)", label, dstID)
	}
	seg := EdgeSegment(srcG, dstG)
	return Seed{
		Label: label, SrcNode: srcID, DstNode: dstID,
		SX: seg.Start.X, SY: seg.Start.Y, SZ: seg.Start.Z,
		EX: seg.End.X, EY: seg.End.Y, EZ: seg.End.Z,
	}, nil
}

func MutualPairs(pairs [][2]string) map[string]map[string]bool {
	hasEdge := make(map[string]bool, len(pairs))
	for _, p := range pairs {
		hasEdge[p[0]+"\x00"+p[1]] = true
	}
	out := map[string]map[string]bool{}
	for _, p := range pairs {
		if !hasEdge[p[1]+"\x00"+p[0]] {
			continue
		}
		if out[p[0]] == nil {
			out[p[0]] = map[string]bool{}
		}
		out[p[0]][p[1]] = true
	}
	return out
}
