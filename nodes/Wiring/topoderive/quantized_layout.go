package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func ComputeQuantizedLayout(spec loadspec.TopoSpec, sphere geom.SceneSphere, centers map[string]wire.Vec3, nodeGeoms map[string]nodegeom.NodeGeom) map[string]quantoffset.QuantizedOffset {
	ids := make(map[string]bool, len(spec.Nodes))
	for _, n := range spec.Nodes {
		ids[n.ID] = true
	}

	prior := make(map[string]quantoffset.QuantizedOffset, len(spec.Nodes))
	for _, n := range spec.Nodes {
		o := quantoffset.QuantizedOffset{}
		if n.StepTheta != nil {
			o.CTheta = *n.StepTheta
		}
		if n.StepPhi != nil {
			o.CPhi = *n.StepPhi
		}
		if n.StepR != nil {
			o.CR = *n.StepR
		}
		prior[n.ID] = o
	}

	measured := quantoffset.MeasureScalars(centers, ids, sphere.Center, prior)
	offsets := make(map[string]quantoffset.QuantizedOffset, len(spec.Nodes))

	exact := make(map[string]bool, len(spec.Nodes))
	for _, n := range spec.Nodes {
		if n.ScenePolarR != nil && n.ScenePolarTheta != nil && n.ScenePolarPhi != nil {
			exact[n.ID] = true
			if off, ok := measured[n.ID]; ok {
				offsets[n.ID] = off
			} else {
				offsets[n.ID] = prior[n.ID]
			}
			continue
		}
		if n.QuantITheta != nil && n.QuantIPhi != nil && n.QuantIR != nil {
			o := quantoffset.QuantizedOffset{
				ITheta: *n.QuantITheta,
				IPhi:   *n.QuantIPhi,
				IR:     *n.QuantIR,
			}
			if n.StepTheta != nil {
				o.CTheta = *n.StepTheta
			}
			if n.StepPhi != nil {
				o.CPhi = *n.StepPhi
			}
			if n.StepR != nil {
				o.CR = *n.StepR
			}
			offsets[n.ID] = o
			continue
		}
		if off, ok := measured[n.ID]; ok {
			offsets[n.ID] = off
			continue
		}
		offsets[n.ID] = prior[n.ID]
	}

	for id, o := range offsets {
		offsets[id] = quantoffset.NormalizeOffset(o)
	}

	derived := quantoffset.DeriveCenters(offsets, sphere.Center)
	for id, pos := range derived {
		if exact[id] {
			continue
		}
		centers[id] = pos
		if g, ok := nodeGeoms[id]; ok {
			nodegeom.SetNodeWorld(&g, pos)
			nodeGeoms[id] = g
		}
	}
	return offsets
}
