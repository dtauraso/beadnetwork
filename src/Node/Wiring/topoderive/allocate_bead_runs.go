package topoderive

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/loadspec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/wire"

	T "github.com/dtauraso/wirefold/src/Trace"
)

func AllocateBeadRuns(spec loadspec.TopoSpec, nodeGeoms map[string]nodegeom.NodeGeom, tr *T.Trace) (
	destRun map[string]*wire.BeadRun,
	edgeRun loadspec.BeadRunRegistry,
	edgeEndpoints map[string]inputcodec.EdgeEndpoints,
) {
	destRun = map[string]*wire.BeadRun{}
	edgeRun = loadspec.BeadRunRegistry{}
	edgeEndpoints = map[string]inputcodec.EdgeEndpoints{}
	for _, e := range spec.Edges {
		destKey := e.Target + "." + e.TargetHandle

		if _, exists := destRun[destKey]; exists {
			panic("AllocateBeadRuns: two edges target " + destKey + " — validateNoFanIn should have rejected this fan-in at parse")
		}

		pw := wire.NewBeadRun()
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
	return destRun, edgeRun, edgeEndpoints
}
