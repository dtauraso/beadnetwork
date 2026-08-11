// reach.go — the reach-radius derive phase and the pure polar-distance math it rides
// on, lifted out of nodes/Wiring/loader_layout.go and nodes/Wiring/broadcast_move.go.
//
// ReachRFromPolar stays exported so nodes/Wiring's own commitNodeMoveLocal (a live-drag
// path that stays in Wiring) can call it too — one-way import (Wiring -> topoderive),
// never the reverse.
package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

// ComputeReachRadii computes each node's REACH radius (max distance from its
// center to any node it outputs to) under the loaded centers — non-rooted layout
// — streamed in NodeGeometry's sphereR field so the TS SphereRing reaches every
// surface node. Computed before newMoveDispatch so each node/edge mover captures
// it in its held geom.
// ComputeReachRadii mutates nodeGeoms IN PLACE (writing each node's ReachR field), so it
// has nothing to return.
func ComputeReachRadii(spec loadspec.TopoSpec, nodeGeoms map[string]nodegeom.NodeGeom) {
	edges := make([]geom.SphereEdge, 0, len(spec.Edges))
	for _, e := range spec.Edges {
		edges = append(edges, geom.SphereEdge{Source: e.Source, Target: e.Target})
	}
	polars := map[string]geom.Polar{}
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

// ReachRFromPolar computes each node's sphere REACH radius (max distance from a node to any
// node it outputs to) under the given polar positions and edge set. Distance is the spherical
// law-of-cosines distance between the two polar positions (polarDist) — no cartesian, no vector
// subtraction. Called by ComputeReachRadii (load path) and by nodes/Wiring's RootMove (live-drag
// path) so the fanned "center" message carries the new reach radius and the ring stays sized
// during a drag.
func ReachRFromPolar(polars map[string]geom.Polar, edges []geom.SphereEdge) map[string]float64 {
	reachR := map[string]float64{}
	for _, e := range edges {
		sp, okS := polars[e.Source]
		tp, okT := polars[e.Target]
		if !okS || !okT {
			continue
		}
		if d := geom.PolarDist(sp, tp); d > reachR[e.Source] {
			reachR[e.Source] = d
		}
	}
	return reachR
}
