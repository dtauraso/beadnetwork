package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/spatial"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"

	T "github.com/dtauraso/wirefold/Trace"
)

func AllocateWires(spec loadspec.TopoSpec, nodeGeoms map[string]nodegeom.NodeGeom, tr *T.Trace) (
	destWire map[string]*wire.PacedWire,
	edgeWire loadspec.WireRegistry,
	edgeEndpoints map[string]inputcodec.EdgeEndpoints,
	edgeSteps map[string]int,
	edgeSegments map[string]spatial.WireSegment,
) {
	destWire = map[string]*wire.PacedWire{}
	edgeWire = loadspec.WireRegistry{}
	edgeEndpoints = map[string]inputcodec.EdgeEndpoints{}
	edgeSteps = map[string]int{}
	edgeSegments = map[string]spatial.WireSegment{}
	for _, e := range spec.Edges {
		destKey := e.Target + "." + e.TargetHandle

		srcG, tgtG := nodeGeoms[e.Source], nodeGeoms[e.Target]
		seg := nodegeom.EdgeSegment(srcG, tgtG)

		dist := nodegeom.NodeWorldPos(srcG).Sub(nodegeom.NodeWorldPos(tgtG)).Length()
		steps := nodegeom.EdgeStepCount(dist, srcG.Kind, tgtG.Kind)
		edgeSteps[e.Label] = steps
		edgeSegments[e.Label] = seg

		if _, exists := destWire[destKey]; exists {
			panic("AllocateWires: two edges target " + destKey + " — validateNoFanIn should have rejected this fan-in at parse")
		}

		pw := wire.NewPacedWire(steps, lattice.DwellTicksPerBead)
		pw.Target = e.Target
		pw.TargetHandle = e.TargetHandle
		pw.SetTrace(tr)
		destWire[destKey] = pw
		edgeWire[e.Label] = pw
		edgeEndpoints[e.Label] = inputcodec.EdgeEndpoints{
			Source: e.Source, Target: e.Target,
			SourceHandle: e.SourceHandle, TargetHandle: e.TargetHandle,
		}
	}
	return destWire, edgeWire, edgeEndpoints, edgeSteps, edgeSegments
}
