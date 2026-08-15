package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func ComputeDragIndices(spec loadspec.TopoSpec) map[string]polarindex.Offset {
	out := make(map[string]polarindex.Offset, len(spec.Nodes))
	for _, n := range spec.Nodes {
		var o polarindex.Offset
		if n.DragIndexPhi != nil {
			o.Phi = *n.DragIndexPhi
		}
		if n.DragIndexTheta != nil {
			o.Theta = *n.DragIndexTheta
		}
		if n.DragIndexR != nil {
			o.R = *n.DragIndexR
		}
		out[n.ID] = o
	}
	return out
}

func ComputeBaseIndices(spec loadspec.TopoSpec, sphere polar.SceneSphere, centers map[string]spatial.Vec3, nodeGeoms map[string]nodegeom.NodeGeom) map[string]polarindex.Index {
	ids := make(map[string]bool, len(spec.Nodes))
	for _, n := range spec.Nodes {
		ids[n.ID] = true
	}

	measured := polarindex.MeasureScalars(centers, ids, sphere.Center, spec.Constants)
	offsets := make(map[string]polarindex.Index, len(spec.Nodes))

	exact := make(map[string]bool, len(spec.Nodes))
	for _, n := range spec.Nodes {
		if n.IndexPhi != nil && n.IndexTheta != nil && n.IndexR != nil {
			offsets[n.ID] = polarindex.Canonical(polarindex.Index{
				Phi:   *n.IndexPhi,
				Theta: *n.IndexTheta,
				R:     *n.IndexR,
			}, spec.Constants)
			continue
		}
		if off, ok := measured[n.ID]; ok {
			offsets[n.ID] = off
			continue
		}
		offsets[n.ID] = polarindex.Index{}
	}

	for id, off := range offsets {
		if exact[id] {
			continue
		}
		if g, ok := nodeGeoms[id]; ok {
			nodegeom.SetNodeWorld(&g, off)
			nodeGeoms[id] = g
			centers[id] = nodegeom.NodeWorldPos(g)
		}
	}
	return offsets
}
