package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

func ComputeReachRadii(spec loadspec.TopoSpec, nodeGeoms map[string]nodegeom.NodeGeom) {
	edges := make([]polar.SphereEdge, 0, len(spec.Edges))
	for _, e := range spec.Edges {
		edges = append(edges, polar.SphereEdge{Source: e.Source, Target: e.Target})
	}
	polars := map[string]polar.Polar{}
	for id, g := range nodeGeoms {
		if g.HasPos {
			polars[id] = g.ScenePolar
		}
	}
	for id, r := range ReachRFromPolar(polars, edges) {
		g := nodeGeoms[id]
		g.ReachR = r
		nodeGeoms[id] = g
	}
}

func ReachRFromPolar(polars map[string]polar.Polar, edges []polar.SphereEdge) map[string]float64 {
	reachR := map[string]float64{}
	for _, e := range edges {
		sp, okS := polars[e.Source]
		tp, okT := polars[e.Target]
		if !okS || !okT {
			continue
		}
		if d := polar.PolarDist(sp, tp); d > reachR[e.Source] {
			reachR[e.Source] = d
		}
	}
	return reachR
}
