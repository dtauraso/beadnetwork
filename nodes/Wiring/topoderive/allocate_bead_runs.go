package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/bead"
	lattice "github.com/dtauraso/wirefold/nodes/bead/lattice"
	"github.com/dtauraso/wirefold/nodes/spatial"

	T "github.com/dtauraso/wirefold/Trace"
)

func AllocateBeadRuns(spec loadspec.TopoSpec, nodeGeoms map[string]nodegeom.NodeGeom, tr *T.Trace) (
	destRun map[string]*bead.BeadRun,
	edgeRun loadspec.BeadRunRegistry,
	edgeEndpoints map[string]inputcodec.EdgeEndpoints,
	edgeSteps map[string]int,
	edgeSegments map[string]spatial.Segment,
) {
	destRun = map[string]*bead.BeadRun{}
	edgeRun = loadspec.BeadRunRegistry{}
	edgeEndpoints = map[string]inputcodec.EdgeEndpoints{}
	edgeSteps = map[string]int{}
	edgeSegments = map[string]spatial.Segment{}
	for _, e := range spec.Edges {
		destKey := e.Target + "." + e.TargetHandle

		srcG, tgtG := nodeGeoms[e.Source], nodeGeoms[e.Target]
		seg := edgegeom.EdgeSegment(srcG, tgtG)

		dist := nodegeom.NodeWorldPos(srcG).Sub(nodegeom.NodeWorldPos(tgtG)).Length()
		steps := edgegeom.EdgeStepCount(dist, srcG.Kind, tgtG.Kind)
		edgeSteps[e.Label] = steps
		edgeSegments[e.Label] = seg

		if _, exists := destRun[destKey]; exists {
			panic("AllocateBeadRuns: two edges target " + destKey + " — validateNoFanIn should have rejected this fan-in at parse")
		}

		pw := bead.NewBeadRun(steps, lattice.DwellTicksPerBead)
		pw.Owner = e.Source
		pw.Edge = e.Label
		pw.Target = e.Target
		pw.TargetHandle = e.TargetHandle
		pw.SetTrace(tr)
		destRun[destKey] = pw
		edgeRun[e.Label] = pw
		edgeEndpoints[e.Label] = inputcodec.EdgeEndpoints{
			Source: e.Source, Target: e.Target,
			SourceHandle: e.SourceHandle, TargetHandle: e.TargetHandle,
		}
	}
	return destRun, edgeRun, edgeEndpoints, edgeSteps, edgeSegments
}
