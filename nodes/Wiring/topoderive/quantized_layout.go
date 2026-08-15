package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func ComputeDragQuantOffsets(spec loadspec.TopoSpec) map[string]quantoffset.QuantizedOffset {
	out := make(map[string]quantoffset.QuantizedOffset, len(spec.Nodes))
	for _, n := range spec.Nodes {
		var o quantoffset.QuantizedOffset
		if n.DragIPhi != nil {
			o.IPhi = *n.DragIPhi
		}
		if n.DragITheta != nil {
			o.ITheta = *n.DragITheta
		}
		if n.DragIR != nil {
			o.IR = *n.DragIR
		}
		out[n.ID] = o
	}
	return out
}

func ComputeQuantizedLayout(spec loadspec.TopoSpec, sphere polar.SceneSphere, centers map[string]spatial.Vec3, nodeGeoms map[string]nodegeom.NodeGeom) map[string]quantoffset.QuantizedOffset {
	ids := make(map[string]bool, len(spec.Nodes))
	for _, n := range spec.Nodes {
		ids[n.ID] = true
	}

	prior := make(map[string]quantoffset.QuantizedOffset, len(spec.Nodes))
	for _, n := range spec.Nodes {
		o := quantoffset.QuantizedOffset{}
		if n.ConstantPhi != nil {
			o.ConstantPhi = *n.ConstantPhi
		}
		if n.ConstantTheta != nil {
			o.ConstantTheta = *n.ConstantTheta
		}
		if n.ConstantR != nil {
			o.ConstantR = *n.ConstantR
		}
		prior[n.ID] = o
	}

	measured := quantoffset.MeasureScalars(centers, ids, sphere.Center, prior)
	offsets := make(map[string]quantoffset.QuantizedOffset, len(spec.Nodes))

	exact := make(map[string]bool, len(spec.Nodes))
	for _, n := range spec.Nodes {
		if n.ScenePolarR != nil && n.ScenePolarPhi != nil && n.ScenePolarTheta != nil {
			exact[n.ID] = true
			if off, ok := measured[n.ID]; ok {
				offsets[n.ID] = off
			} else {
				offsets[n.ID] = prior[n.ID]
			}
			continue
		}
		if n.IPhi != nil && n.ITheta != nil && n.IR != nil {
			o := quantoffset.QuantizedOffset{
				IPhi:   *n.IPhi,
				ITheta: *n.ITheta,
				IR:     *n.IR,
			}
			if n.ConstantPhi != nil {
				o.ConstantPhi = *n.ConstantPhi
			}
			if n.ConstantTheta != nil {
				o.ConstantTheta = *n.ConstantTheta
			}
			if n.ConstantR != nil {
				o.ConstantR = *n.ConstantR
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
